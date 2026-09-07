package dtos

import "github.com/bayuanugerah/insurance-core-api/internal/models"

type CreateProductRequest struct {
	Name             string               `json:"name"`
	Slug             string               `json:"slug"`
	Category         string               `json:"category"`
	Status           string               `json:"status,omitempty"`
	ShortDescription string               `json:"short_description"`
	Description      string               `json:"description"`
	TargetCustomer   string               `json:"target_customer"`
	MinSumAssured    int64                `json:"min_sum_assured"`
	MaxSumAssured    int64                `json:"max_sum_assured"`
	MinPaymentTerm   int                  `json:"min_payment_term"`
	MaxPaymentTerm   int                  `json:"max_payment_term"`
	StartingPremium  int64                `json:"starting_premium"`
	NonMCULimit      int64                `json:"non_mcu_limit,omitempty"`
	Benefits         []string             `json:"benefits"`
	Exclusions       []string             `json:"exclusions"`
	IsFeatured       bool                 `json:"is_featured"`
	PricingRules     *models.PricingRules `json:"pricing_rules,omitempty"`
}

type UpdateProductRequest struct {
	Name             *string              `json:"name,omitempty"`
	Slug             *string              `json:"slug,omitempty"`
	Category         *string              `json:"category,omitempty"`
	Status           *string              `json:"status,omitempty"`
	ShortDescription *string              `json:"short_description,omitempty"`
	Description      *string              `json:"description,omitempty"`
	TargetCustomer   *string              `json:"target_customer,omitempty"`
	MinSumAssured    *int64               `json:"min_sum_assured,omitempty"`
	MaxSumAssured    *int64               `json:"max_sum_assured,omitempty"`
	MinPaymentTerm   *int                 `json:"min_payment_term,omitempty"`
	MaxPaymentTerm   *int                 `json:"max_payment_term,omitempty"`
	StartingPremium  *int64               `json:"starting_premium,omitempty"`
	NonMCULimit      *int64               `json:"non_mcu_limit,omitempty"`
	Benefits         *[]string            `json:"benefits,omitempty"`
	Exclusions       *[]string            `json:"exclusions,omitempty"`
	IsFeatured       *bool                `json:"is_featured,omitempty"`
	PricingRules     *models.PricingRules `json:"pricing_rules,omitempty"`
}

type UpdateProductStatusRequest struct {
	Status string `json:"status"`
}

type ProductManagementMetricsResponse struct {
	TotalProducts          int     `json:"total_products"`
	ActiveProducts         int     `json:"active_products"`
	DraftProducts          int     `json:"draft_products"`
	ArchivedProducts       int     `json:"archived_products"`
	TotalActivePolicies    int     `json:"total_active_policies"`
	TotalGWPVolume         int64   `json:"total_gwp_volume"`
	TotalGWPVolumeFormatted string  `json:"total_gwp_volume_formatted"`
	AverageLossRatio       float64 `json:"average_loss_ratio"`
}
