package services

import (
	"context"
	"errors"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
	"github.com/google/uuid"
)

type ProductKnowledgeSyncer interface {
	SyncProductKnowledge(ctx context.Context, product models.Product) error
	RemoveProductKnowledge(ctx context.Context, productSlug string) error
}

type ProductService struct {
	productRepository     repositories.ProductRepository
	pricingRuleRepository repositories.PricingRuleRepository
	knowledgeSyncer       ProductKnowledgeSyncer
}

func NewProductService(productRepository repositories.ProductRepository, pricingRuleRepository ...repositories.PricingRuleRepository) *ProductService {
	var prRepo repositories.PricingRuleRepository
	if len(pricingRuleRepository) > 0 {
		prRepo = pricingRuleRepository[0]
	}
	return &ProductService{productRepository: productRepository, pricingRuleRepository: prRepo}
}

func (service *ProductService) WithKnowledgeSyncer(syncer ProductKnowledgeSyncer) *ProductService {
	service.knowledgeSyncer = syncer
	return service
}

func (service *ProductService) ListProducts(ctx context.Context, input dtos.ProductListQuery) ([]models.Product, error) {
	return service.productRepository.FindAll(ctx, repositories.ProductFilter{
		Category:   strings.TrimSpace(input.Category),
		Status:     strings.TrimSpace(input.Status),
		IsFeatured: input.IsFeatured,
		Limit:      input.Limit,
		Search:     strings.TrimSpace(input.Search),
	})
}

func (service *ProductService) GetProductBySlug(ctx context.Context, slug string) (models.Product, error) {
	return service.productRepository.FindBySlug(ctx, slug)
}

func (service *ProductService) GetProductByID(ctx context.Context, id string) (models.Product, error) {
	return service.productRepository.FindByID(ctx, id)
}

func (service *ProductService) CreateProduct(ctx context.Context, rawReq dtos.CreateProductRequest) (models.Product, error) {
	req, err := validations.ValidateCreateProductRequest(rawReq)
	if err != nil {
		return models.Product{}, err
	}

	// Verify slug uniqueness
	if _, err := service.productRepository.FindBySlug(ctx, req.Slug); err == nil {
		return models.Product{}, constants.ErrProductSlugAlreadyExistsError
	}

	product := models.Product{
		ID:               uuid.New().String(),
		Name:             req.Name,
		Slug:             req.Slug,
		Category:         models.ProductCategory(req.Category),
		Status:           models.ProductStatus(req.Status),
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
		TargetCustomer:   req.TargetCustomer,
		MinSumAssured:    req.MinSumAssured,
		MaxSumAssured:    req.MaxSumAssured,
		MinPaymentTerm:   req.MinPaymentTerm,
		MaxPaymentTerm:   req.MaxPaymentTerm,
		StartingPremium:  req.StartingPremium,
		NonMCULimit:      req.NonMCULimit,
		Benefits:         req.Benefits,
		Exclusions:       req.Exclusions,
		IsFeatured:       req.IsFeatured,
	}
	if req.PricingRules != nil {
		product.PricingRules = *req.PricingRules
	}

	if err := service.productRepository.Create(ctx, &product); err != nil {
		return models.Product{}, err
	}

	service.syncKnowledge(ctx, product)

	return product, nil
}

func (service *ProductService) UpdateProduct(ctx context.Context, id string, rawReq dtos.UpdateProductRequest) (models.Product, error) {
	req, err := validations.ValidateUpdateProductRequest(rawReq)
	if err != nil {
		return models.Product{}, err
	}

	product, err := service.productRepository.FindByID(ctx, id)
	if err != nil {
		product, err = service.productRepository.FindBySlug(ctx, id)
		if err != nil {
			return models.Product{}, constants.ErrProductNotFoundError
		}
	}

	if req.Slug != nil && *req.Slug != product.Slug {
		if existing, err := service.productRepository.FindBySlug(ctx, *req.Slug); err == nil && existing.ID != product.ID {
			return models.Product{}, constants.ErrProductSlugAlreadyExistsError
		}
		product.Slug = *req.Slug
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Category != nil {
		product.Category = models.ProductCategory(*req.Category)
	}
	if req.Status != nil {
		product.Status = models.ProductStatus(*req.Status)
	}
	if req.ShortDescription != nil {
		product.ShortDescription = *req.ShortDescription
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.TargetCustomer != nil {
		product.TargetCustomer = *req.TargetCustomer
	}
	if req.MinSumAssured != nil {
		product.MinSumAssured = *req.MinSumAssured
	}
	if req.MaxSumAssured != nil {
		product.MaxSumAssured = *req.MaxSumAssured
	}
	if req.MinPaymentTerm != nil {
		product.MinPaymentTerm = *req.MinPaymentTerm
	}
	if req.MaxPaymentTerm != nil {
		product.MaxPaymentTerm = *req.MaxPaymentTerm
	}
	if req.StartingPremium != nil {
		product.StartingPremium = *req.StartingPremium
	}
	if req.NonMCULimit != nil {
		product.NonMCULimit = *req.NonMCULimit
	}
	if req.Benefits != nil {
		product.Benefits = *req.Benefits
	}
	if req.Exclusions != nil {
		product.Exclusions = *req.Exclusions
	}
	if req.IsFeatured != nil {
		product.IsFeatured = *req.IsFeatured
	}
	if req.PricingRules != nil {
		product.PricingRules = *req.PricingRules
	}

	if product.MaxSumAssured < product.MinSumAssured {
		return models.Product{}, errors.New(constants.ErrProductMaxSumAssuredInvalid)
	}
	if product.MaxPaymentTerm < product.MinPaymentTerm {
		return models.Product{}, errors.New(constants.ErrProductMaxPaymentTermInvalid)
	}

	if err := service.productRepository.Update(ctx, &product); err != nil {
		return models.Product{}, err
	}

	service.syncKnowledge(ctx, product)

	return product, nil
}

func (service *ProductService) UpdateProductStatus(ctx context.Context, id string, status models.ProductStatus) (models.Product, error) {
	product, err := service.productRepository.FindByID(ctx, id)
	if err != nil {
		product, err = service.productRepository.FindBySlug(ctx, id)
		if err != nil {
			return models.Product{}, constants.ErrProductNotFoundError
		}
	}

	if err := service.productRepository.UpdateStatus(ctx, product.ID, status); err != nil {
		return models.Product{}, err
	}

	product.Status = status
	service.syncKnowledge(ctx, product)

	return product, nil
}

func (service *ProductService) ToggleProductStatus(ctx context.Context, id string) (models.Product, error) {
	product, err := service.productRepository.FindByID(ctx, id)
	if err != nil {
		product, err = service.productRepository.FindBySlug(ctx, id)
		if err != nil {
			return models.Product{}, constants.ErrProductNotFoundError
		}
	}

	targetStatus := models.ProductStatusActive
	if product.Status == models.ProductStatusActive {
		targetStatus = models.ProductStatusDraft
	}

	return service.UpdateProductStatus(ctx, product.ID, targetStatus)
}

func (service *ProductService) DeleteProduct(ctx context.Context, id string) error {
	product, err := service.productRepository.FindByID(ctx, id)
	if err != nil {
		product, err = service.productRepository.FindBySlug(ctx, id)
		if err != nil {
			return constants.ErrProductNotFoundError
		}
	}

	hasApps, err := service.productRepository.HasApplications(ctx, product.ID)
	if err != nil {
		return err
	}
	if hasApps {
		return constants.ErrProductHasApplicationsError
	}

	if err := service.productRepository.Delete(ctx, product.ID); err != nil {
		return err
	}

	service.removeKnowledge(ctx, product.Slug)

	return nil
}

func (service *ProductService) GetProductManagementMetrics(ctx context.Context) (dtos.ProductManagementMetricsResponse, error) {
	return service.productRepository.GetMetrics(ctx)
}

func (service *ProductService) UpdatePricingRules(ctx context.Context, slug string, rules []models.ProductPricingRule) error {
	product, err := service.productRepository.FindBySlug(ctx, slug)
	if err != nil {
		return constants.ErrProductNotFoundError
	}

	if service.pricingRuleRepository == nil {
		return errors.New("pricing rule repository not configured")
	}

	return service.pricingRuleRepository.SaveBatch(ctx, product.ID, rules)
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

	if product.Status == models.ProductStatusDraft || product.Status == models.ProductStatusArchived {
		return dtos.ProductQuote{}, repositories.ErrProductNotFound
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

	if product.Category == constants.CategoryVehicle {
		genderFactor = 1.0
		smokerFactor = 1.0
		healthFactor = 1.0
	}

	if !validFactor(ageFactor) {
		return dtos.ProductQuoteBreakdown{}, constants.QuoteAgeOutOfRangeError
	}
	if !validFactor(baseRate) || !validFactor(genderFactor) || !validFactor(smokerFactor) || !validFactor(occupationFactor) || !validFactor(healthFactor) || !validFactor(frequencyLoading) || !validFactor(termFactor) {
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
			ruleCodeLower := strings.ToLower(rule.RuleCode)
			if product.Category == constants.CategoryVehicle && (strings.Contains(ruleCodeLower, "gender") || strings.Contains(ruleCodeLower, "smoker") || strings.Contains(ruleCodeLower, "health") || strings.Contains(ruleCodeLower, "illness")) {
				continue
			}
			val := findAnswerForRule(rule, input)
			factor := parseMultiplierFactor(rule.Factors, val)
			if factor <= 0 {
				factor = 1.0
			}
			dynamicMultiplier *= factor

			factorsBreakdown = append(factorsBreakdown, dtos.ProductQuoteFactorItem{
				RuleCode: rule.RuleCode,
				RuleName: rule.RuleName,
				Factor:   factor,
			})
		}
	}

	if !validFactor(ageFactor) {
		return dtos.ProductQuote{}, constants.QuoteAgeOutOfRangeError
	}

	if !validFactor(baseRate) || !validFactor(termFactor) || !validFactor(frequencyLoading) || !validFactor(dynamicMultiplier) {
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
	var list []any
	if raw, ok := factors["brackets"]; ok {
		if l, ok := raw.([]any); ok {
			list = l
		}
	} else if raw, ok := factors["age_factors"]; ok {
		if l, ok := raw.([]any); ok {
			list = l
		}
	}
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

func (service *ProductService) syncKnowledge(ctx context.Context, product models.Product) {
	if service.knowledgeSyncer == nil {
		return
	}
	if product.Status == models.ProductStatusArchived {
		if err := service.knowledgeSyncer.RemoveProductKnowledge(ctx, product.Slug); err != nil {
			log.Printf("[ProductService] warning: failed to remove knowledge for %s: %v", product.Slug, err)
		}
	} else {
		if err := service.knowledgeSyncer.SyncProductKnowledge(ctx, product); err != nil {
			log.Printf("[ProductService] warning: failed to sync knowledge for %s: %v", product.Slug, err)
		}
	}
}

func (service *ProductService) removeKnowledge(ctx context.Context, slug string) {
	if service.knowledgeSyncer == nil {
		return
	}
	if err := service.knowledgeSyncer.RemoveProductKnowledge(ctx, slug); err != nil {
		log.Printf("[ProductService] warning: failed to remove knowledge for %s: %v", slug, err)
	}
}
