package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductAnalytics struct {
	ID                uint      `gorm:"primaryKey"`
	ProductID         uint      `gorm:"uniqueIndex"`
	LastAnalyzedAt    time.Time
	PriceChangeCount  int
	StockChangeCount  int
	FavoriteCount     int
	ViewCount        int
	AddToCartCount   int
	OrderCount       int
	PopularityScore  float64
	UpdatePriority   int       // Higher number means higher priority
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type PriceHistory struct {
	ID         uint      `gorm:"primaryKey"`
	VariantID  uint      `gorm:"index"`
	OldPrice   float64
	NewPrice   float64
	ChangedAt  time.Time
	CreatedAt  time.Time
}

type StockHistory struct {
	ID           uint      `gorm:"primaryKey"`
	VariantID    uint      `gorm:"index"`
	OldQuantity  int
	NewQuantity  int
	ChangedAt    time.Time
	CreatedAt    time.Time
}

type UpdatePriority struct {
	ID          uint      `gorm:"primaryKey"`
	ProductID   uint      `gorm:"uniqueIndex"`
	Priority    int       // 1: Normal, 2: Favorited, 3: High-demand
	LastUpdated time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BeforeCreate will set the timestamps
func (pa *ProductAnalytics) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	pa.CreatedAt = now
	pa.UpdatedAt = now
	pa.LastAnalyzedAt = now
	return nil
}

// BeforeUpdate will update the timestamps
func (pa *ProductAnalytics) BeforeUpdate(tx *gorm.DB) error {
	pa.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate will set the timestamps
func (ph *PriceHistory) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	ph.CreatedAt = now
	ph.ChangedAt = now
	return nil
}

// BeforeCreate will set the timestamps
func (sh *StockHistory) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	sh.CreatedAt = now
	sh.ChangedAt = now
	return nil
}

// BeforeCreate will set the timestamps
func (up *UpdatePriority) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	up.CreatedAt = now
	up.UpdatedAt = now
	up.LastUpdated = now
	return nil
}

// BeforeUpdate will update the timestamps
func (up *UpdatePriority) BeforeUpdate(tx *gorm.DB) error {
	up.UpdatedAt = time.Now()
	return nil
}
