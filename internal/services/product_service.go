package services

import (
	"context"
	"math"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type ProductService struct {
	productRepository repositories.ProductRepository
}

func NewProductService(productRepository repositories.ProductRepository) *ProductService {
	return &ProductService{productRepository: productRepository}
}

func (service *ProductService) ListProducts(ctx context.Context, input dtos.ProductListQuery) ([]models.Product, error) {
	return service.productRepository.FindAll(ctx, repositories.ProductFilter{
		Category:   strings.TrimSpace(input.Category),
		IsFeatured: input.IsFeatured,
		Limit:      input.Limit,
	})
}

func (service *ProductService) GetProductBySlug(ctx context.Context, slug string) (models.Product, error) {
	return service.productRepository.FindBySlug(ctx, slug)
}

func (service *ProductService) CreateProductQuote(ctx context.Context, slug string, input dtos.CreateProductQuoteInput) (dtos.ProductQuote, error) {
	product, err := service.productRepository.FindBySlug(ctx, slug)
	if err != nil {
		return dtos.ProductQuote{}, err
	}

	if input.SumAssured < product.MinSumAssured || input.SumAssured > product.MaxSumAssured {
		return dtos.ProductQuote{}, constants.QuoteSumAssuredOutOfRangeError
	}

	if input.PaymentTerm < product.MinPaymentTerm || input.PaymentTerm > product.MaxPaymentTerm {
		return dtos.ProductQuote{}, constants.QuotePaymentTermOutOfRangeError
	}

	breakdown, err := buildQuoteBreakdown(product, input)
	if err != nil {
		return dtos.ProductQuote{}, err
	}

	annualPremium := float64(input.SumAssured) * breakdown.BaseRate * breakdown.AgeFactor * breakdown.GenderFactor * breakdown.SmokerFactor * breakdown.OccupationFactor * breakdown.HealthFactor * breakdown.TermFactor
	frequencyDivisor, ok := paymentFrequencyDivisor(input.PaymentFrequency)
	if !ok {
		return dtos.ProductQuote{}, constants.QuotePricingRulesInvalidError
	}
	periodicPremium := annualPremium / frequencyDivisor * breakdown.FrequencyLoading

	return dtos.ProductQuote{
		ProductID:              product.ID,
		ProductName:            product.Name,
		Currency:               constants.CurrencyIDR,
		Age:                    input.Age,
		Gender:                 input.Gender,
		SumAssured:             input.SumAssured,
		PaymentTerm:            input.PaymentTerm,
		PaymentFrequency:       input.PaymentFrequency,
		EstimatedPremium:       roundUpToNearestThousand(periodicPremium),
		EstimatedAnnualPremium: roundUpToNearestThousand(annualPremium),
		Breakdown:              breakdown,
		Notes:                  constants.ProductQuoteNotes(),
	}, nil
}

func buildQuoteBreakdown(product models.Product, input dtos.CreateProductQuoteInput) (dtos.ProductQuoteBreakdown, error) {
	rules := product.PricingRules
	baseRate := rules.BaseRate
	ageFactor := findAgeFactor(rules.AgeFactors, input.Age)
	genderFactor := findRuleFactor(rules.GenderFactors, input.Gender)
	smokerFactor := findRuleFactor(rules.SmokerFactors, input.Smoker)
	occupationFactor := findRuleFactor(rules.OccupationFactors, input.OccupationClass)
	healthFactor := findRuleFactor(rules.HealthFactors, input.HealthRisk)
	frequencyLoading := findRuleFactor(rules.FrequencyLoading, input.PaymentFrequency)
	termFactor := calculateTermFactor(product.MinPaymentTerm, input.PaymentTerm)

	if !validFactor(baseRate) || !validFactor(ageFactor) || !validFactor(genderFactor) || !validFactor(smokerFactor) || !validFactor(occupationFactor) || !validFactor(healthFactor) || !validFactor(frequencyLoading) || !validFactor(termFactor) {
		return dtos.ProductQuoteBreakdown{}, constants.QuotePricingRulesInvalidError
	}

	return dtos.ProductQuoteBreakdown{
		BaseRate:         baseRate,
		AgeFactor:        ageFactor,
		GenderFactor:     genderFactor,
		SmokerFactor:     smokerFactor,
		OccupationFactor: occupationFactor,
		HealthFactor:     healthFactor,
		TermFactor:       termFactor,
		FrequencyLoading: frequencyLoading,
	}, nil
}

func validFactor(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func findAgeFactor(ageFactors []models.AgeFactor, age int) float64 {
	for _, ageFactor := range ageFactors {
		if age >= ageFactor.MinAge && age <= ageFactor.MaxAge {
			return ageFactor.Factor
		}
	}

	return 0
}

func findRuleFactor(factors map[string]float64, key string) float64 {
	if factors == nil {
		return 0
	}

	return factors[key]
}

func calculateTermFactor(minPaymentTerm int, paymentTerm int) float64 {
	termDelta := paymentTerm - minPaymentTerm
	if termDelta < 0 {
		termDelta = 0
	}

	return 1 + (float64(termDelta) * 0.01)
}

func paymentFrequencyDivisor(paymentFrequency string) (float64, bool) {
	switch paymentFrequency {
	case constants.PaymentFrequencyAnnual:
		return 1, true
	case constants.PaymentFrequencySemiAnnual:
		return 2, true
	case constants.PaymentFrequencyQuarterly:
		return 4, true
	case constants.PaymentFrequencyMonthly:
		return 12, true
	default:
		return 0, false
	}
}

func roundUpToNearestThousand(value float64) int64 {
	return int64(math.Ceil(value/1000) * 1000)
}
