package models

import (
	"time"
)

type Category struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	ParentID  *uint     `gorm:"index"`
	ExternalID string   `gorm:"size:255"`
	Slug      string    `gorm:"size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CategoryCrawlStatus struct {
	ID            uint      `gorm:"primaryKey"`
	CategoryID    uint      `gorm:"index"`
	LastCrawledAt time.Time
	Status        string    `gorm:"size:50"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Product struct {
	ID              uint      `gorm:"primaryKey"`
	ExternalID      string    `gorm:"size:255;uniqueIndex;not null"`
	Name            string    `gorm:"size:500;not null"`
	CategoryID      *uint     `gorm:"index"`
	BrandID         *uint     `gorm:"index"`
	Description     string    `gorm:"type:text"`
	RatingScore     float64
	FavoriteCount   int       `gorm:"default:0"`
	CommentCount    int       `gorm:"default:0"`
	ViewCount       int       `gorm:"default:0"`
	AddToCartCount  int       `gorm:"default:0"`
	OrderCount      int       `gorm:"default:0"`
	SizeRecommendation string `gorm:"size:255"`
	EstimatedDelivery string  `gorm:"size:255"`
	IsActive        bool      `gorm:"default:true"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastCrawledAt   *time.Time
}

type ProductImage struct {
	ID         uint   `gorm:"primaryKey"`
	ProductID  uint   `gorm:"index"`
	URL        string `gorm:"size:500;not null"`
	SortOrder  int    
	IsVideo    bool   `gorm:"default:false"`
	CreatedAt  time.Time
}

type ProductVariant struct {
	ID                uint      `gorm:"primaryKey"`
	ProductID         uint      `gorm:"index"`
	SKU               string    `gorm:"size:255"`
	ExternalVariantID string    `gorm:"size:255"`
	Color             string    `gorm:"size:100"`
	Size              string    `gorm:"size:100"`
	Price             float64
	OriginalPrice     *float64
	StockQuantity     int       `gorm:"default:0"`
	IsActive          bool      `gorm:"default:true"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ProductAttribute struct {
	ID            uint   `gorm:"primaryKey"`
	ProductID     uint   `gorm:"index"`
	AttributeName string `gorm:"size:255;not null"`
	AttributeValue string `gorm:"type:text;not null"`
	CreatedAt     time.Time
}

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"size:255;not null"`
	Email     string    `gorm:"size:255;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserFavorite struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	ProductID uint      `gorm:"index"`
	CreatedAt time.Time
}

type PriceHistory struct {
	ID        uint      `gorm:"primaryKey"`
	VariantID uint      `gorm:"index"`
	OldPrice  float64
	NewPrice  float64
	ChangedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type StockHistory struct {
	ID          uint      `gorm:"primaryKey"`
	VariantID   uint      `gorm:"index"`
	OldQuantity int
	NewQuantity int
	ChangedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type Notification struct {
	ID              uint      `gorm:"primaryKey"`
	UserID          uint      `gorm:"index"`
	ProductID       uint      `gorm:"index"`
	VariantID       *uint     `gorm:"index"`
	NotificationType string   `gorm:"size:50;not null"`
	Message         string    `gorm:"type:text;not null"`
	IsRead          bool      `gorm:"default:false"`
	CreatedAt       time.Time
}