package dtos

type ApplicationMetricsDTO struct {
	Total        int64   `json:"total"`
	Draft        int64   `json:"draft"`
	Submitted    int64   `json:"submitted"`
	UnderReview  int64   `json:"under_review"`
	Approved     int64   `json:"approved"`
	Rejected     int64   `json:"rejected"`
	ApprovalRate float64 `json:"approval_rate"`
}

type PremiumMetricsDTO struct {
	TotalWrittenPremium int64 `json:"total_written_premium"`
	ActivePoliciesCount int64 `json:"active_policies_count"`
}

type UnderwritingMetricsDTO struct {
	PendingReviewCount int64   `json:"pending_review_count"`
	AverageSLAMinutes  float64 `json:"average_sla_minutes"`
	SLATargetMinutes   int     `json:"sla_target_minutes"`
	SLAComplianceRate  float64 `json:"sla_compliance_rate"`
}

type ProductPerformanceMetricDTO struct {
	ProductID           string  `json:"product_id"`
	ProductSlug         string  `json:"product_slug"`
	ProductName         string  `json:"product_name"`
	Category            string  `json:"category"`
	ActivePoliciesCount int64   `json:"active_policies_count"`
	TotalPremium        int64   `json:"total_premium"`
	LossRatio           float64 `json:"loss_ratio"`
}

type AdminMetricsResponse struct {
	Applications ApplicationMetricsDTO         `json:"applications"`
	Premiums     PremiumMetricsDTO             `json:"premiums"`
	Underwriting UnderwritingMetricsDTO        `json:"underwriting"`
	TopProducts  []ProductPerformanceMetricDTO `json:"top_products"`
}
