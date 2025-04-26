package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/faisaloncode/ecommerce-crawler/notification/models"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type NotificationService struct {
	db *gorm.DB
	kafkaReader *kafka.Reader
}

type PriceChangeEvent struct {
	ProductID string  `json:"product_id"`
	OldPrice  float64 `json:"old_price"`
	NewPrice  float64 `json:"new_price"`
}

type StockChangeEvent struct {
	ProductID string `json:"product_id"`
	InStock   bool   `json:"in_stock"`
}

func NewNotificationService(db *gorm.DB, kafkaBroker string) *NotificationService {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   "product-updates",
		GroupID: "notification-service",
	})

	return &NotificationService{
		db:          db,
		kafkaReader: reader,
	}
}

func (s *NotificationService) Start() {
	go s.consumeKafkaMessages()
}

func (s *NotificationService) consumeKafkaMessages() {
	for {
		msg, err := s.kafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading Kafka message: %v", err)
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		eventType, ok := event["type"].(string)
		if !ok {
			log.Printf("Invalid event type")
			continue
		}

		switch eventType {
		case "price_change":
			s.handlePriceChange(event)
		case "stock_change":
			s.handleStockChange(event)
		}
	}
}

func (s *NotificationService) handlePriceChange(event map[string]interface{}) {
	data := event["data"].(map[string]interface{})
	priceChange := PriceChangeEvent{
		ProductID: data["product_id"].(string),
		OldPrice:  data["old_price"].(float64),
		NewPrice:  data["new_price"].(float64),
	}

	var prefs []models.NotificationPreference
	s.db.Where("product_id = ? AND min_price >= ? AND max_price <= ?",
		priceChange.ProductID, priceChange.NewPrice, priceChange.OldPrice).Find(&prefs)

	for _, pref := range prefs {
		notification := models.Notification{
			UserID:    pref.UserID,
			ProductID: pref.ProductID,
			Type:      models.PriceDropNotification,
			Message:   "Price dropped from " + formatPrice(priceChange.OldPrice) + " to " + formatPrice(priceChange.NewPrice),
		}
		s.db.Create(&notification)
	}
}

func (s *NotificationService) handleStockChange(event map[string]interface{}) {
	data := event["data"].(map[string]interface{})
	stockChange := StockChangeEvent{
		ProductID: data["product_id"].(string),
		InStock:   data["in_stock"].(bool),
	}

	var prefs []models.NotificationPreference
	s.db.Where("product_id = ? AND notify_stock = ?", stockChange.ProductID, true).Find(&prefs)

	for _, pref := range prefs {
		status := "back in stock"
		if !stockChange.InStock {
			status = "out of stock"
		}

		notification := models.Notification{
			UserID:    pref.UserID,
			ProductID: pref.ProductID,
			Type:      models.StockChangeNotification,
			Message:   "Product is now " + status,
		}
		s.db.Create(&notification)
	}
}

func formatPrice(price float64) string {
	return "₺" + formatFloat(price)
}

func formatFloat(num float64) string {
	return string(rune(num))
}
