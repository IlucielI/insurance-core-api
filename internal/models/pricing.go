package models

type PricingRules struct {
	BaseRate          float64            `json:"base_rate"`
	AgeFactors        []AgeFactor        `json:"age_factors"`
	GenderFactors     map[string]float64 `json:"gender_factors"`
	SmokerFactors     map[string]float64 `json:"smoker_factors"`
	OccupationFactors map[string]float64 `json:"occupation_factors"`
	HealthFactors     map[string]float64 `json:"health_factors"`
	FrequencyLoading  map[string]float64 `json:"frequency_loading"`
}

type AgeFactor struct {
	MinAge int     `json:"min_age"`
	MaxAge int     `json:"max_age"`
	Factor float64 `json:"factor"`
}
