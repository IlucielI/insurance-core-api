package constants

const (
	ProductQueryCategory = "category"
	ProductQueryFeatured = "featured"
	ProductQueryLimit    = "limit"

	MaxProductListLimit = 50
	MinQuoteAge         = 18
	MaxQuoteAge         = 60
)

const (
	GenderMale   = "male"
	GenderFemale = "female"
)

const (
	SmokerYes = "yes"
	SmokerNo  = "no"
)

const (
	OccupationLow      = "low"
	OccupationStandard = "standard"
	OccupationHigh     = "high"
)

const (
	HealthRiskLow    = "low"
	HealthRiskMedium = "medium"
	HealthRiskHigh   = "high"
)

const (
	PaymentFrequencyAnnual     = "annual"
	PaymentFrequencySemiAnnual = "semi_annual"
	PaymentFrequencyQuarterly  = "quarterly"
	PaymentFrequencyMonthly    = "monthly"
)

const (
	CurrencyIDR = "IDR"
)

func ProductQuoteNotes() []string {
	return []string{
		"This is an indicative quote, not a final offer.",
		"Final premium may change after underwriting review.",
	}
}
