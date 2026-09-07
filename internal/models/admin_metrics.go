package models

type ApplicationStatusCount struct {
	Status ApplicationStatus `gorm:"column:status"`
	Count  int64             `gorm:"column:count"`
}

type ProductVolumeAggregate struct {
	ProductID           string `gorm:"column:product_id"`
	ProductSlug         string `gorm:"column:product_slug"`
	ProductName         string `gorm:"column:product_name"`
	Category            string `gorm:"column:category"`
	ActivePoliciesCount int64  `gorm:"column:active_policies_count"`
	TotalPremium        int64  `gorm:"column:total_premium"`
}

type SLAMetricsAggregate struct {
	TotalReviewed     int64
	TotalWithinSLA    int64
	TotalDurationMins float64
}

type AdminMetrics struct {
	StatusCounts        []ApplicationStatusCount
	ActivePoliciesCount int64
	TotalWrittenPremium int64
	SLAMetrics          SLAMetricsAggregate
	TopProducts         []ProductVolumeAggregate
}
