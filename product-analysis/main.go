package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/faisaloncode/ecommerce-crawler/product-analysis/config"
	"github.com/faisaloncode/ecommerce-crawler/product-analysis/models"
	pb "github.com/faisaloncode/ecommerce-crawler/product-analysis/proto"
	"github.com/faisaloncode/ecommerce-crawler/product-analysis/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(
		&models.ProductAnalytics{},
		&models.PriceHistory{},
		&models.StockHistory{},
		&models.UpdatePriority{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	productAnalysisService := service.NewProductAnalysisService(db)
	pb.RegisterProductAnalysisServiceServer(grpcServer, productAnalysisService)

	// Start HTTP server for health checks
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy"}`)) 
		})

		httpPort := 8082 // Use a different port for HTTP
		log.Printf("Starting HTTP server on port %d", httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Printf("Starting product analysis service on port %d", cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return db, nil
}
