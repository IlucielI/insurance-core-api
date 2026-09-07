package models

import (
	"time"
)

type NotificationSeverity string

const (
	NotificationSeverityInfo     NotificationSeverity = "INFO"
	NotificationSeverityWarning  NotificationSeverity = "WARNING"
	NotificationSeverityCritical NotificationSeverity = "CRITICAL"
	NotificationSeveritySuccess  NotificationSeverity = "SUCCESS"
)

type NotificationCategory string

const (
	NotificationCategoryUnderwriting NotificationCategory = "underwriting"
	NotificationCategorySystem       NotificationCategory = "system"
	NotificationCategoryKnowledge    NotificationCategory = "knowledge"
	NotificationCategoryPolicy       NotificationCategory = "policy"
)

type Notification struct {
	ID        string               `gorm:"primaryKey;size:64" json:"id"`
	Type      string               `gorm:"size:32;not null" json:"type"`
	Category  NotificationCategory `gorm:"size:32;not null" json:"category"`
	Severity  NotificationSeverity `gorm:"size:16;not null;default:'INFO'" json:"severity"`
	Title     string               `gorm:"size:255;not null" json:"title"`
	Message   string               `gorm:"type:text;not null" json:"message"`
	Link      string               `gorm:"size:255;not null;default:''" json:"link"`
	IsRead    bool                 `gorm:"not null;default:false" json:"is_read"`
	CreatedAt time.Time            `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	ReadAt    *time.Time           `json:"read_at,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}
