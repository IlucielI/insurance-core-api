package models

import "time"

type ProductCategory string

const (
	ProductCategoryLife    ProductCategory = "life"
	ProductCategoryHealth  ProductCategory = "health"
	ProductCategoryVehicle ProductCategory = "vehicle"
)

type Product struct {
	ID               string          `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Name             string          `gorm:"type:varchar(120);not null" json:"name"`
	Slug             string          `gorm:"type:varchar(120);uniqueIndex;not null" json:"slug"`
	Category         ProductCategory `gorm:"type:varchar(32);not null;index" json:"category"`
	ShortDescription string          `gorm:"type:varchar(255);not null" json:"short_description"`
	Description      string          `gorm:"type:text;not null" json:"description"`
	TargetCustomer   string          `gorm:"type:varchar(255);not null" json:"target_customer"`
	MinSumAssured    int64           `gorm:"not null" json:"min_sum_assured"`
	MaxSumAssured    int64           `gorm:"not null" json:"max_sum_assured"`
	MinPaymentTerm   int             `gorm:"not null" json:"min_payment_term"`
	MaxPaymentTerm   int             `gorm:"not null" json:"max_payment_term"`
	StartingPremium  int64           `gorm:"not null" json:"starting_premium"`
	Benefits         []string        `gorm:"type:jsonb;serializer:json" json:"benefits"`
	Exclusions       []string        `gorm:"type:jsonb;serializer:json" json:"exclusions"`
	IsFeatured       bool            `gorm:"not null;default:false;index" json:"is_featured"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (Product) TableName() string {
	return "products"
}
