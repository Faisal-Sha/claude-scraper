package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/faisaloncode/ecommerce-crawler/crawler/config"
	"github.com/faisaloncode/ecommerce-crawler/crawler/crawler"
	// "github.com/faisaloncode/ecommerce-crawler/crawler/models"
	"github.com/faisaloncode/ecommerce-crawler/crawler/proto"
	"github.com/faisaloncode/ecommerce-crawler/crawler/scraper"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize DB connection
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run database migrations
	// migrateDB(db)

	// Initialize category scraper
	categoryScraper := scraper.NewCategoryScraper(db, cfg)

	// Initialize crawler service with category scraper
	crawlerService := crawler.NewCrawlerService(db, nil, categoryScraper) // nil for product analysis client as crawler doesn't need to make requests

	// Start the crawler service
	go crawlerService.StartScheduler()

	// Setup gRPC server
	lis, err := net.Listen("tcp", cfg.CrawlerServiceAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register the crawler service
	proto.RegisterCrawlerServiceServer(grpcServer, crawlerService)

	log.Printf("Starting crawler service on %s", cfg.CrawlerServiceAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// func migrateDB(db *gorm.DB) {
// 	// Run migrations specific to the crawler service
// 	err := db.AutoMigrate(
// 		&models.Category{},
// 		&models.Product{},
// 	)
// 	if err != nil {
// 		log.Fatalf("Failed to run migrations: %v", err)
// 	}
// 	log.Println("Database migrations completed")
// }
