package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/faisaloncode/ecommerce-crawler/product-analysis/models"
	pb "github.com/faisaloncode/ecommerce-crawler/product-analysis/proto"
)

type ProductAnalysisService struct {
	pb.UnimplementedProductAnalysisServiceServer
	db *gorm.DB
}

func NewProductAnalysisService(db *gorm.DB) *ProductAnalysisService {
	return &ProductAnalysisService{
		db: db,
	}
}

func (s *ProductAnalysisService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: "healthy"}, nil
}

func (s *ProductAnalysisService) AnalyzeProduct(ctx context.Context, req *pb.AnalyzeProductRequest) (*pb.AnalyzeProductResponse, error) {
	productID := req.Product.Id
	var notifications []string

	// Get or create product analytics
	var analytics models.ProductAnalytics
	productIDUint, err := strconv.ParseUint(productID, 10, 64)
if err != nil {
	return nil, fmt.Errorf("invalid product ID: %v", err)
}
result := s.db.FirstOrCreate(&analytics, models.ProductAnalytics{ProductID: uint(productIDUint)})
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get/create product analytics: %v", result.Error)
	}

	// Update analytics
	analytics.ViewCount = int(req.Product.ViewCount)
	analytics.FavoriteCount = int(req.Product.FavoriteCount)
	analytics.AddToCartCount = int(req.Product.AddToCartCount)
	analytics.OrderCount = int(req.Product.OrderCount)

	// Calculate popularity score (simple weighted average)
	analytics.PopularityScore = calculatePopularityScore(
		float64(req.Product.ViewCount),
		float64(req.Product.FavoriteCount),
		float64(req.Product.AddToCartCount),
		float64(req.Product.OrderCount),
	)

	// Check variants for price and stock changes
	for _, variant := range req.Product.Variants {
		notifications = append(notifications, s.analyzeVariant(variant)...)
	}

	// Update last analyzed time
	analytics.LastAnalyzedAt = time.Now()
	s.db.Save(&analytics)

	return &pb.AnalyzeProductResponse{
		Status:        "success",
		Notifications: notifications,
	}, nil
}

func (s *ProductAnalysisService) UpdateProductPriority(ctx context.Context, req *pb.UpdateProductPriorityRequest) (*pb.UpdateProductPriorityResponse, error) {
	var priority models.UpdatePriority
	productIDUint, err := strconv.ParseUint(req.ProductId, 10, 64)
if err != nil {
	return nil, fmt.Errorf("invalid product ID: %v", err)
}
result := s.db.FirstOrCreate(&priority, models.UpdatePriority{ProductID: uint(productIDUint)})
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get/create update priority: %v", result.Error)
	}

	// Set priority based on favorited status
	if req.IsFavorited {
		priority.Priority = 2 // Favorited products get higher priority
	} else {
		priority.Priority = 1 // Normal priority
	}

	priority.LastUpdated = time.Now()
	s.db.Save(&priority)

	return &pb.UpdateProductPriorityResponse{Status: "success"}, nil
}

func (s *ProductAnalysisService) GetProductAnalytics(ctx context.Context, req *pb.GetProductAnalyticsRequest) (*pb.GetProductAnalyticsResponse, error) {
	var analytics models.ProductAnalytics
	result := s.db.First(&analytics, "product_id = ?", req.ProductId)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get product analytics: %v", result.Error)
	}

	// Get price history
	var priceHistory []models.PriceHistory
	s.db.Where("variant_id IN (SELECT id FROM product_variants WHERE product_id = ?)", req.ProductId).
		Order("changed_at DESC").
		Limit(10).
		Find(&priceHistory)

	pbPriceHistory := make([]*pb.PriceHistory, len(priceHistory))
	for i, ph := range priceHistory {
		pbPriceHistory[i] = &pb.PriceHistory{
			VariantId: fmt.Sprint(ph.VariantID),
			OldPrice:  float32(ph.OldPrice),
			NewPrice:  float32(ph.NewPrice),
			ChangedAt: ph.ChangedAt.Format(time.RFC3339),
		}
	}

	return &pb.GetProductAnalyticsResponse{
		PriceTrend:        calculatePriceTrend(priceHistory),
		StockTrend:        0, // TODO: Implement stock trend calculation
		FavoriteCountTrend: int32(analytics.FavoriteCount),
		PopularityScore:    float32(analytics.PopularityScore),
		PriceHistory:       pbPriceHistory,
	}, nil
}

func (s *ProductAnalysisService) analyzeVariant(variant *pb.ProductVariant) []string {
	var notifications []string

	// Check for price changes
	var lastPriceHistory models.PriceHistory
	result := s.db.Where("variant_id = ?", variant.Id).
		Order("changed_at DESC").
		First(&lastPriceHistory)

	variantIDUint, err := strconv.ParseUint(variant.Id, 10, 64)
if err != nil {
	return []string{fmt.Sprintf("error:invalid_variant_id=%s", variant.Id)}
}

if result.Error == nil {
		if float64(variant.Price) != lastPriceHistory.NewPrice {
			// Price has changed
			priceHistory := models.PriceHistory{
				VariantID: uint(variantIDUint),
				OldPrice:  lastPriceHistory.NewPrice,
				NewPrice:  float64(variant.Price),
			}
			s.db.Create(&priceHistory)

			if variant.Price < float32(lastPriceHistory.NewPrice) {
				notifications = append(notifications, fmt.Sprintf("price_drop:variant_id=%s:old_price=%.2f:new_price=%.2f",
					variant.Id, lastPriceHistory.NewPrice, variant.Price))
			}
		}
	} else {
		// First price record
		priceHistory := models.PriceHistory{
			VariantID: uint(variantIDUint),
			OldPrice:  float64(variant.Price),
			NewPrice:  float64(variant.Price),
		}
		s.db.Create(&priceHistory)
	}

	// Check for stock changes
	var lastStockHistory models.StockHistory
	result = s.db.Where("variant_id = ?", variantIDUint).
		Order("changed_at DESC").
		First(&lastStockHistory)

	if result.Error == nil {
		if int(variant.StockQuantity) != lastStockHistory.NewQuantity {
			// Stock has changed
			stockHistory := models.StockHistory{
				VariantID:   uint(variantIDUint),
				OldQuantity: lastStockHistory.NewQuantity,
				NewQuantity: int(variant.StockQuantity),
			}
			s.db.Create(&stockHistory)

			if variant.StockQuantity == 0 {
				notifications = append(notifications, fmt.Sprintf("out_of_stock:variant_id=%s", variant.Id))
			}
		}
	} else {
		// First stock record
		stockHistory := models.StockHistory{
			VariantID:   uint(variantIDUint),
			OldQuantity: int(variant.StockQuantity),
			NewQuantity: int(variant.StockQuantity),
		}
		s.db.Create(&stockHistory)
	}

	return notifications
}

func calculatePopularityScore(views, favorites, addToCarts, orders float64) float64 {
	// Simple weighted average
	return (views*0.1 + favorites*0.2 + addToCarts*0.3 + orders*0.4) / 1000
}

func calculatePriceTrend(history []models.PriceHistory) float32 {
	if len(history) < 2 {
		return 0
	}

	// Calculate percentage change between first and last price
	firstPrice := history[len(history)-1].NewPrice
	lastPrice := history[0].NewPrice
	return float32((lastPrice - firstPrice) / firstPrice * 100)
}
