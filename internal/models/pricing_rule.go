package models

import "time"

type ProductPricingRule struct {
	ID         string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	ProductID  string         `gorm:"type:varchar(64);not null;index" json:"product_id"`
	RuleCode   string         `gorm:"type:varchar(64);not null" json:"rule_code"`
	RuleName   string         `gorm:"type:varchar(120);not null" json:"rule_name"`
	RuleType   string         `gorm:"type:varchar(32);not null" json:"rule_type"`
	Factors    map[string]any `gorm:"type:jsonb;serializer:json" json:"factors"`
	IsActive   bool           `gorm:"not null;default:true" json:"is_active"`
	OrderIndex int            `gorm:"not null;default:0" json:"order_index"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func (ProductPricingRule) TableName() string {
	return "product_pricing_rules"
}
