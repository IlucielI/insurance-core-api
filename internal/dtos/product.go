package dtos

type ProductListQuery struct {
	Category   string
	IsFeatured *bool
	Limit      int
}

type ProductQuoteRequest struct {
	Age              int    `json:"age"`
	Gender           string `json:"gender"`
	SumAssured       int64  `json:"sum_assured"`
	PaymentTerm      int    `json:"payment_term"`
	PaymentFrequency string `json:"payment_frequency"`
	Smoker           string `json:"smoker"`
	OccupationClass  string `json:"occupation_class"`
	HealthRisk       string `json:"health_risk"`
}

type CreateProductQuoteInput struct {
	Age              int
	Gender           string
	SumAssured       int64
	PaymentTerm      int
	PaymentFrequency string
	Smoker           string
	OccupationClass  string
	HealthRisk       string
}

type ProductQuote struct {
	ProductID              string                `json:"product_id"`
	ProductName            string                `json:"product_name"`
	Currency               string                `json:"currency"`
	Age                    int                   `json:"age"`
	Gender                 string                `json:"gender"`
	SumAssured             int64                 `json:"sum_assured"`
	PaymentTerm            int                   `json:"payment_term"`
	PaymentFrequency       string                `json:"payment_frequency"`
	EstimatedPremium       int64                 `json:"estimated_premium"`
	EstimatedAnnualPremium int64                 `json:"estimated_annual_premium"`
	Breakdown              ProductQuoteBreakdown `json:"breakdown"`
	Notes                  []string              `json:"notes"`
}

type ProductQuoteBreakdown struct {
	BaseRate         float64 `json:"base_rate"`
	AgeFactor        float64 `json:"age_factor"`
	GenderFactor     float64 `json:"gender_factor"`
	SmokerFactor     float64 `json:"smoker_factor"`
	OccupationFactor float64 `json:"occupation_factor"`
	HealthFactor     float64 `json:"health_factor"`
	TermFactor       float64 `json:"term_factor"`
	FrequencyLoading float64 `json:"frequency_loading"`
}
