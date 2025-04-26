package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/faisaloncode/ecommerce-crawler/notification/config"
	"github.com/faisaloncode/ecommerce-crawler/notification/models"
	"github.com/faisaloncode/ecommerce-crawler/notification/service"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	// Drop tables if they exist
	db.Migrator().DropTable(&models.Notification{})
	db.Migrator().DropTable(&models.NotificationPreference{})

	// Run migrations
	err = db.AutoMigrate(
		&models.Notification{},
		&models.NotificationPreference{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize notification service
	notificationService := service.NewNotificationService(db, cfg.KafkaBroker)
	notificationService.Start()

	// Initialize Echo server
	e := echo.New()

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})

	// Get notifications endpoint
	e.GET("/notifications/:user_id", func(c echo.Context) error {
		var notifications []models.Notification
		userID := c.Param("user_id")
		
		result := db.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications)
		if result.Error != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": result.Error.Error()})
		}

		return c.JSON(http.StatusOK, notifications)
	})

	// Mark notification as read endpoint
	e.PUT("/notifications/:id/read", func(c echo.Context) error {
		id := c.Param("id")
		
		result := db.Model(&models.Notification{}).Where("id = ?", id).Update("is_read", true)
		if result.Error != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": result.Error.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "success"})
	})

	// Set notification preferences endpoint
	e.POST("/preferences", func(c echo.Context) error {
		pref := new(models.NotificationPreference)
		if err := c.Bind(pref); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		result := db.Create(pref)
		if result.Error != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": result.Error.Error()})
		}

		return c.JSON(http.StatusCreated, pref)
	})

	// Start server
	log.Printf("Starting notification service on port %s", cfg.ServerPort)
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
