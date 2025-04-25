package crawler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/faisaloncode/ecommerce-crawler/crawler/models"
	pb "github.com/faisaloncode/ecommerce-crawler/crawler/proto"
	"github.com/faisaloncode/ecommerce-crawler/crawler/scraper"
)

type CrawlerService struct {
	pb.UnimplementedCrawlerServiceServer
	db                    *gorm.DB
	productAnalysisClient pb.ProductAnalysisServiceClient
	categoryScraper      *scraper.CategoryScraper
	httpClient            *http.Client
	baseURL               string
}

func NewCrawlerService(db *gorm.DB, productAnalysisClient pb.ProductAnalysisServiceClient, categoryScraper *scraper.CategoryScraper) *CrawlerService {
	return &CrawlerService{
		db:                    db,
		productAnalysisClient: productAnalysisClient,
		categoryScraper:      categoryScraper,
		httpClient:            &http.Client{Timeout: 10 * time.Second},
		baseURL:               "https://example.com",
	}
}

func (s *CrawlerService) StartScheduler() {
	log.Println("Starting crawler scheduler...")
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.CrawlAllCategories()
		}
	}
}

// Health implements the Health RPC method
func (s *CrawlerService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: "healthy"}, nil
}

// ListCategories implements the ListCategories RPC method
func (s *CrawlerService) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	var categories []models.Category
	result := s.db.Find(&categories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch categories: %v", result.Error)
	}

	pbCategories := make([]*pb.Category, len(categories))
	for i, cat := range categories {
		pbCategories[i] = &pb.Category{
			Id:   fmt.Sprint(cat.ID),
			Name: cat.Name,
			Url:  fmt.Sprintf("%s/categories/%s", s.baseURL, cat.Slug),
		}
	}

	return &pb.ListCategoriesResponse{Categories: pbCategories}, nil
}

// RefreshCategories implements the RefreshCategories RPC method
func (s *CrawlerService) RefreshCategories(ctx context.Context, req *pb.RefreshCategoriesRequest) (*pb.RefreshCategoriesResponse, error) {
	go s.CrawlAllCategories()
	return &pb.RefreshCategoriesResponse{Status: "refresh started"}, nil
}

// ListProducts implements the ListProducts RPC method
func (s *CrawlerService) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	var products []models.Product
	result := s.db.Where("category_id = ?", req.CategoryId).Offset(int(req.Page * req.PerPage)).Limit(int(req.PerPage)).Find(&products)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch products: %v", result.Error)
	}

	var total int64
	s.db.Model(&models.Product{}).Where("category_id = ?", req.CategoryId).Count(&total)

	pbProducts := make([]*pb.Product, len(products))
	for i, prod := range products {
		pbProducts[i] = &pb.Product{
			Id:          fmt.Sprint(prod.ID),
			Name:        prod.Name,
			Description: prod.Description,
			CategoryId:  fmt.Sprint(prod.CategoryID),
		}
	}

	return &pb.ListProductsResponse{Products: pbProducts, Total: int32(total)}, nil
}

// GetProduct implements the GetProduct RPC method
func (s *CrawlerService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	var product models.Product
	result := s.db.First(&product, req.Id)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch product: %v", result.Error)
	}

	pbProduct := &pb.Product{
		Id:          fmt.Sprint(product.ID),
		Name:        product.Name,
		Description: product.Description,
		CategoryId:  fmt.Sprint(product.CategoryID),
	}

	return &pb.GetProductResponse{Product: pbProduct}, nil
}

func (s *CrawlerService) CrawlAllCategories() {
	var categories []models.Category
	
	// Get all leaf categories (those without children)
	subQuery := s.db.Model(&models.Category{}).Select("parent_id")
	result := s.db.Where("id NOT IN (?)", subQuery).Find(&categories)
	
	if result.Error != nil {
		log.Printf("Error fetching categories: %v", result.Error)
		return
	}

	log.Printf("Found %d categories to crawl", len(categories))
	for _, category := range categories {
		go s.CrawlCategory(fmt.Sprintf("%d", category.ID))
		
		// Add a small delay to avoid overwhelming the server
		time.Sleep(500 * time.Millisecond)
	}
}

func (s *CrawlerService) CrawlCategory(categoryID string) {
	log.Printf("Crawling category ID: %s", categoryID)

	// Update crawl status
	var status models.CategoryCrawlStatus
	s.db.Where("category_id = ?", categoryID).FirstOrCreate(&status)
	status.Status = "in_progress"
	status.LastCrawledAt = time.Now()
	s.db.Save(&status)

	// Get category details
	var category models.Category
	if err := s.db.First(&category, categoryID).Error; err != nil {
		log.Printf("Category not found: %v", err)
		status.Status = "failed"
		s.db.Save(&status)
		return
	}

	var products []models.Product

	// Mock implementation for development
	// Replace with real implementation in production
	products = append(products, models.Product{
		ExternalID: "12345",
		Name:       "Mock Product 1",
	})
	products = append(products, models.Product{
		ExternalID: "67890",
		Name:       "Mock Product 2",
	})

	log.Printf("Found %d products for category %s", len(products), categoryID)

	// Process each product
	for _, product := range products {
		// Mock implementation for development
		productData := &pb.ProductData{
			ExternalId:  product.ExternalID,
			Name:        product.Name,
			Description: "This is a mock product for testing",
			Price:       99.99,
			Stock:       10,
			IsActive:    true,
			CategoryId:  categoryID,
			BrandId:     "45",
			SellerId:    "67",
			Images: []*pb.ProductImage{
				{Url: "https://example.com/image1.jpg", IsVideo: false},
			},
			Variants: []*pb.ProductVariant{
				{
					ExternalVariantId: "variant-" + product.ExternalID + "-1",
					Color:             "Red",
					Size:              "M",
					Price:             99.99,
					Stock:             5,
				},
				{
					ExternalVariantId: "variant-" + product.ExternalID + "-2",
					Color:             "Blue",
					Size:              "L",
					Price:             109.99,
					Stock:             5,
				},
			},
		}

		// Send to Product Analysis Service via gRPC
		s.sendProductToAnalysis(productData)
		
		// Add a small delay to avoid overwhelming the server
		time.Sleep(200 * time.Millisecond)
	}

	// Update crawl status
	status.Status = "completed"
	s.db.Save(&status)
	log.Printf("Completed crawling category ID: %s", categoryID)
}

func (s *CrawlerService) sendProductToAnalysis(productData *pb.ProductData) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := s.productAnalysisClient.AnalyzeProduct(ctx, productData)
	if err != nil {
		log.Printf("Failed to send product to analysis service: %v", err)
		return
	}

	log.Printf("Product sent to analysis service. Response: %v", response.Status)
}