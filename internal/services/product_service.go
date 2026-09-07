package services

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type ProductService struct {
	productRepository     repositories.ProductRepository
	pricingRuleRepository repositories.PricingRuleRepository
}

func NewProductService(productRepository repositories.ProductRepository, pricingRuleRepository ...repositories.PricingRuleRepository) *ProductService {
	var prRepo repositories.PricingRuleRepository
	if len(pricingRuleRepository) > 0 {
		prRepo = pricingRuleRepository[0]
	}
	return &ProductService{productRepository: productRepository, pricingRuleRepository: prRepo}
}

func (service *ProductService) ListProducts(ctx context.Context, input dtos.ProductListQuery) ([]models.Product, error) {
	return service.productRepository.FindAll(ctx, repositories.ProductFilter{
		Category:   strings.TrimSpace(input.Category),
		IsFeatured: input.IsFeatured,
		Limit:      input.Limit,
		Search:     strings.TrimSpace(input.Search),
	})
}

func (service *ProductService) GetProductBySlug(ctx context.Context, slug string) (models.Product, error) {
	return service.productRepository.FindBySlug(ctx, slug)
}

func (service *ProductService) GetPricingRules(ctx context.Context, slug string) ([]models.ProductPricingRule, error) {
	if service.pricingRuleRepository == nil {
		return nil, nil
	}
	return service.pricingRuleRepository.FindByProductSlug(ctx, slug)
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

	// 1. Try dynamic pricing rules from database
	if service.pricingRuleRepository != nil {
		rules, err := service.pricingRuleRepository.FindByProductID(ctx, product.ID)
		if err == nil && len(rules) > 0 {
			return service.calculateDynamicQuote(product, rules, input)
		}
	}

	// 2. Fallback to product JSON pricing rules
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
		ProductSlug:            product.Slug,
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

func (service *ProductService) calculateDynamicQuote(
	product models.Product,
	rules []models.ProductPricingRule,
	input dtos.CreateProductQuoteInput,
) (dtos.ProductQuote, error) {
	baseRate := 0.0
	ageFactor := 0.0
	frequencyLoading := 1.0
	termFactor := calculateTermFactor(product.MinPaymentTerm, input.PaymentTerm)

	dynamicMultiplier := 1.0
	factorsBreakdown := make([]dtos.ProductQuoteFactorItem, 0, len(rules))

	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}

		switch rule.RuleType {
		case "base_rate":
			baseRate = parseFactorFloat(rule.Factors, "rate")
		case "bracket":
			ageFactor = parseBracketFactor(rule.Factors, input.Age)
		case "frequency_loading":
			frequencyLoading = parseFactorFloat(rule.Factors, input.PaymentFrequency)
		case "multiplier_map":
			val := findAnswerForRule(rule, input)
			factor := parseMultiplierFactor(rule.Factors, val)
			dynamicMultiplier *= factor

			factorsBreakdown = append(factorsBreakdown, dtos.ProductQuoteFactorItem{
				RuleCode: rule.RuleCode,
				RuleName: rule.RuleName,
				Factor:   factor,
			})
		}
	}

	if !validFactor(baseRate) || !validFactor(ageFactor) || !validFactor(termFactor) || !validFactor(frequencyLoading) || !validFactor(dynamicMultiplier) {
		return dtos.ProductQuote{}, constants.QuotePricingRulesInvalidError
	}

	annualPremium := float64(input.SumAssured) * baseRate * ageFactor * termFactor * dynamicMultiplier
	frequencyDivisor, ok := paymentFrequencyDivisor(input.PaymentFrequency)
	if !ok {
		return dtos.ProductQuote{}, constants.QuotePricingRulesInvalidError
	}
	periodicPremium := annualPremium / frequencyDivisor * frequencyLoading

	breakdown := dtos.ProductQuoteBreakdown{
		BaseRate:         baseRate,
		AgeFactor:        ageFactor,
		TermFactor:       termFactor,
		FrequencyLoading: frequencyLoading,
		Factors:          factorsBreakdown,
	}

	for _, item := range factorsBreakdown {
		switch strings.ToLower(item.RuleCode) {
		case "gender":
			breakdown.GenderFactor = item.Factor
		case "smoker":
			breakdown.SmokerFactor = item.Factor
		case "occupation_class", "occupation":
			breakdown.OccupationFactor = item.Factor
		case "health_risk":
			breakdown.HealthFactor = item.Factor
		}
	}

	return dtos.ProductQuote{
		ProductID:              product.ID,
		ProductName:            product.Name,
		ProductSlug:            product.Slug,
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

func parseFactorFloat(factors map[string]any, key string) float64 {
	keyLower := strings.ToLower(strings.TrimSpace(key))
	for k, v := range factors {
		if strings.EqualFold(k, keyLower) {
			switch val := v.(type) {
			case float64:
				return val
			case float32:
				return float64(val)
			case int:
				return float64(val)
			case int64:
				return float64(val)
			case string:
				if parsed, err := strconv.ParseFloat(strings.TrimSpace(val), 64); err == nil {
					return parsed
				}
			}
		}
	}
	return 0
}

func parseBracketFactor(factors map[string]any, age int) float64 {
	if raw, ok := factors["brackets"]; ok {
		if list, ok := raw.([]any); ok {
			for _, item := range list {
				if m, ok := item.(map[string]any); ok {
					minAge := int(parseFactorFloat(m, "min_age"))
					maxAge := int(parseFactorFloat(m, "max_age"))
					factor := parseFactorFloat(m, "factor")
					if age >= minAge && age <= maxAge {
						return factor
					}
				}
			}
		}
	}
	return 0
}

func parseMultiplierFactor(factors map[string]any, key string) float64 {
	keyLower := strings.ToLower(strings.TrimSpace(key))
	for k, v := range factors {
		if strings.EqualFold(k, keyLower) {
			switch val := v.(type) {
			case float64:
				return val
			case float32:
				return float64(val)
			case int:
				return float64(val)
			case int64:
				return float64(val)
			case string:
				if parsed, err := strconv.ParseFloat(strings.TrimSpace(val), 64); err == nil {
					return parsed
				}
			}
		}
	}
	return 1.0
}

func findAnswerForRule(rule models.ProductPricingRule, input dtos.CreateProductQuoteInput) string {
	for _, a := range input.Answers {
		if a.RuleID != "" && strings.EqualFold(a.RuleID, rule.ID) {
			return a.Value
		}
		if a.RuleCode != "" && strings.EqualFold(a.RuleCode, rule.RuleCode) {
			return a.Value
		}
	}

	if strings.EqualFold(rule.RuleCode, "gender") {
		return input.Gender
	}

	return ""
}
