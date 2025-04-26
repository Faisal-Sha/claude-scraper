package models

import (
	"time"

	"gorm.io/gorm"
)

type NotificationType string

const (
	PriceDropNotification NotificationType = "PRICE_DROP"
	StockChangeNotification NotificationType = "STOCK_CHANGE"
)

type Notification struct {
	gorm.Model
	UserID      string          `json:"user_id" gorm:"column:user_id;type:varchar(100);index"`
	ProductID   string          `json:"product_id" gorm:"column:product_id;type:varchar(100);index"`
	Type        NotificationType `json:"type" gorm:"column:type;type:varchar(50)"`
	Message     string          `json:"message" gorm:"column:message;type:text"`
	IsRead      bool            `json:"is_read" gorm:"column:is_read;default:false"`
	CreatedAt   time.Time       `json:"created_at" gorm:"column:created_at"`
}

type NotificationPreference struct {
	gorm.Model
	UserID      string    `json:"user_id" gorm:"column:user_id;type:varchar(100);index"`
	ProductID   string    `json:"product_id" gorm:"column:product_id;type:varchar(100);index"`
	MinPrice    float64   `json:"min_price" gorm:"column:min_price;type:decimal(10,2)"`
	MaxPrice    float64   `json:"max_price" gorm:"column:max_price;type:decimal(10,2)"`
	NotifyStock bool      `json:"notify_stock" gorm:"column:notify_stock"`
}
