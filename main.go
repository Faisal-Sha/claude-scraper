package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"


	pb "github.com/faisaloncode/ecommerce-crawler/crawler/proto"
)

type APIServer struct {
	crawlerClient  pb.CrawlerServiceClient
	analysisClient pb.ProductAnalysisServiceClient
}

func (api *APIServer) healthCheck(c echo.Context) error {
	// Check crawler service health
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := api.crawlerClient.Health(ctx, &pb.HealthRequest{})
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  fmt.Sprintf("crawler service unavailable: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}

func (api *APIServer) listCategories(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := api.crawlerClient.ListCategories(ctx, &pb.ListCategoriesRequest{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp.Categories)
}

func (api *APIServer) refreshCategories(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := api.crawlerClient.RefreshCategories(ctx, &pb.RefreshCategoriesRequest{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "refresh started"})
}

func (api *APIServer) listProducts(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := api.crawlerClient.ListProducts(ctx, &pb.ListProductsRequest{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp.Products)
}

func (api *APIServer) getProduct(c echo.Context) error {
	id := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := api.crawlerClient.GetProduct(ctx, &pb.GetProductRequest{Id: id})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp.Product)
}

func main() {
	// Set service addresses
	crawlerAddr := "localhost:50051"
	analysisAddr := "localhost:50052"

	// Initialize gRPC clients
	crawlerConn, err := initGRPCClient(crawlerAddr)
	if err != nil {
		log.Printf("Warning: Failed to connect to crawler service: %v", err)
	}
	defer crawlerConn.Close()

	analysisConn, err := initGRPCClient(analysisAddr)
	if err != nil {
		log.Printf("Warning: Failed to connect to product analysis service: %v", err)
	}
	defer analysisConn.Close()

	// Initialize API server
	api := &APIServer{
		crawlerClient:  pb.NewCrawlerServiceClient(crawlerConn),
		analysisClient: pb.NewProductAnalysisServiceClient(analysisConn),
	}

	// Setup Echo server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Health check endpoint
	e.GET("/health", api.healthCheck)

	// Category endpoints
	e.GET("/categories", api.listCategories)
	e.POST("/categories/refresh", api.refreshCategories)

	// Start API server
	port := "8082"
	e.Logger.Fatal(e.Start(":" + port))
}

func initGRPCClient(addr string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
}
