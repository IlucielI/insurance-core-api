package models

import (
	"time"
)

type AuditSeverity string

const (
	AuditSeveritySuccess AuditSeverity = "SUCCESS"
	AuditSeverityWarning AuditSeverity = "WARNING"
	AuditSeverityFailed  AuditSeverity = "FAILED"
)

type AuditCategory string

const (
	AuditCategoryUnderwriting AuditCategory = "underwriting"
	AuditCategoryProduct      AuditCategory = "product"
	AuditCategoryKnowledge    AuditCategory = "knowledge"
	AuditCategoryAuth         AuditCategory = "auth"
	AuditCategorySystem       AuditCategory = "system"
)

type AuditLog struct {
	ID             string        `gorm:"primaryKey;size:64" json:"id"`
	Timestamp      time.Time     `gorm:"not null;default:CURRENT_TIMESTAMP" json:"timestamp"`
	ActorName      string        `gorm:"size:128;not null" json:"actor_name"`
	ActorRole      string        `gorm:"size:64;not null" json:"actor_role"`
	Action         string        `gorm:"size:64;not null" json:"action"`
	Category       AuditCategory `gorm:"size:64;not null" json:"category"`
	TargetResource string        `gorm:"size:128;not null" json:"target_resource"`
	IPAddress      string        `gorm:"size:45;not null;default:'127.0.0.1'" json:"ip_address"`
	Status         AuditSeverity `gorm:"size:16;not null;default:'SUCCESS'" json:"status"`
	Details        string        `gorm:"type:jsonb;not null;default:'{}'" json:"details"`
	Hash           string        `gorm:"size:64;not null" json:"hash"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
