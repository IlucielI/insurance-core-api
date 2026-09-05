package models

import "time"

type ApplicationReviewCheckType string

type ApplicationReviewCheckStatus string

const (
	ApplicationReviewCheckTypeIdentityVerified  ApplicationReviewCheckType = "identity_verified"
	ApplicationReviewCheckTypeIncomeVerified    ApplicationReviewCheckType = "income_verified"
	ApplicationReviewCheckTypeDocumentsComplete ApplicationReviewCheckType = "documents_complete"
	ApplicationReviewCheckTypeMedicalRequired   ApplicationReviewCheckType = "medical_required"
)

const (
	ApplicationReviewCheckStatusPending   ApplicationReviewCheckStatus = "pending"
	ApplicationReviewCheckStatusPassed    ApplicationReviewCheckStatus = "passed"
	ApplicationReviewCheckStatusFailed    ApplicationReviewCheckStatus = "failed"
	ApplicationReviewCheckStatusNotNeeded ApplicationReviewCheckStatus = "not_needed"
)

type ApplicationReviewCheck struct {
	ID            string                       `gorm:"primaryKey;type:varchar(64)" json:"id"`
	ApplicationID string                       `gorm:"type:varchar(64);not null;index;uniqueIndex:idx_application_review_check_unique" json:"application_id"`
	CheckType     ApplicationReviewCheckType   `gorm:"type:varchar(64);not null;uniqueIndex:idx_application_review_check_unique" json:"check_type"`
	Status        ApplicationReviewCheckStatus `gorm:"type:varchar(32);not null;index" json:"status"`
	Notes         string                       `gorm:"type:varchar(500)" json:"notes,omitempty"`
	ReviewedBy    string                       `gorm:"type:varchar(120)" json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time                   `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time                    `json:"created_at"`
	UpdatedAt     time.Time                    `json:"updated_at"`
}

func (ApplicationReviewCheck) TableName() string {
	return "application_review_checks"
}
