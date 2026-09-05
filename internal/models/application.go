package models

import "time"

type ApplicationStatus string

const (
	ApplicationStatusSubmitted  ApplicationStatus = "submitted"
	ApplicationStatusUnderReview ApplicationStatus = "under_review"
	ApplicationStatusApproved   ApplicationStatus = "approved"
	ApplicationStatusRejected    ApplicationStatus = "rejected"
)

type Application struct {
	ID                string            `gorm:"primaryKey;type:varchar(64)" json:"id"`
	ProductID         string            `gorm:"type:varchar(64);not null;index" json:"product_id"`
	Product           Product           `gorm:"foreignKey:ProductID" json:"product"`
	FullName          string            `gorm:"type:varchar(120);not null" json:"full_name"`
	Email             string            `gorm:"type:varchar(255);not null" json:"email"`
	Phone             string            `gorm:"type:varchar(32);not null" json:"phone"`
	Age               int               `gorm:"not null" json:"age"`
	Gender            string            `gorm:"type:varchar(16);not null" json:"gender"`
	SumAssured        int64             `gorm:"not null" json:"sum_assured"`
	PaymentTerm       int               `gorm:"not null" json:"payment_term"`
	PaymentFrequency  string            `gorm:"type:varchar(16);not null" json:"payment_frequency"`
	Smoker            string            `gorm:"type:varchar(8);not null" json:"smoker"`
	OccupationClass   string            `gorm:"type:varchar(16);not null" json:"occupation_class"`
	HealthRisk        string            `gorm:"type:varchar(16);not null" json:"health_risk"`
	Premium           int64             `gorm:"not null" json:"premium"`
	Status            ApplicationStatus `gorm:"type:varchar(32);not null;index" json:"status"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

func (Application) TableName() string { return "applications" }
