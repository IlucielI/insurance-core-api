package validations

import (
	"errors"
	"strconv"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func ValidateProductListQuery(categoryValue string, featuredValue string, limitValue string) (dtos.ProductListQuery, error) {
	query := dtos.ProductListQuery{
		Category: strings.TrimSpace(categoryValue),
	}

	if query.Category != "" && validation.Validate(query.Category, validation.In(
		models.ProductCategoryLife,
		models.ProductCategoryHealth,
		models.ProductCategoryVehicle,
	)) != nil {
		return dtos.ProductListQuery{}, errors.New(constants.ErrProductCategoryInvalid)
	}

	featuredValue = strings.TrimSpace(featuredValue)
	if featuredValue != "" {
		isFeatured, err := strconv.ParseBool(featuredValue)
		if err != nil {
			return dtos.ProductListQuery{}, errors.New(constants.ErrProductFeaturedInvalid)
		}

		query.IsFeatured = &isFeatured
	}

	limitValue = strings.TrimSpace(limitValue)
	if limitValue != "" {
		limit, err := strconv.Atoi(limitValue)
		if err != nil || limit < 1 {
			return dtos.ProductListQuery{}, errors.New(constants.ErrProductLimitInvalid)
		}

		if limit > constants.MaxProductListLimit {
			return dtos.ProductListQuery{}, errors.New(constants.ErrProductLimitTooHigh)
		}

		query.Limit = limit
	}

	return query, nil
}

func ValidateProductSlug(slug string) (string, error) {
	slug = strings.TrimSpace(slug)
	if validation.Validate(slug, validation.Required) != nil {
		return "", errors.New(constants.ErrProductSlugRequired)
	}

	return slug, nil
}

func ValidateProductQuoteRequest(request dtos.ProductQuoteRequest) (dtos.ProductQuoteRequest, error) {
	request.Gender = strings.TrimSpace(request.Gender)
	request.PaymentFrequency = strings.TrimSpace(request.PaymentFrequency)
	request.Smoker = strings.TrimSpace(request.Smoker)
	request.OccupationClass = strings.TrimSpace(request.OccupationClass)
	request.HealthRisk = strings.TrimSpace(request.HealthRisk)

	if validation.Validate(request.Age, validation.Min(constants.MinQuoteAge), validation.Max(constants.MaxQuoteAge)) != nil {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuoteAgeInvalid)
	}

	if !contains(request.Gender, constants.GenderMale, constants.GenderFemale) {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuoteGenderInvalid)
	}

	if validation.Validate(request.SumAssured, validation.Min(int64(1))) != nil {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuoteSumAssuredInvalid)
	}

	if validation.Validate(request.PaymentTerm, validation.Min(1)) != nil {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuotePaymentTermInvalid)
	}

	if !contains(request.PaymentFrequency, constants.PaymentFrequencyAnnual, constants.PaymentFrequencySemiAnnual, constants.PaymentFrequencyQuarterly, constants.PaymentFrequencyMonthly) {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuotePaymentFrequencyInvalid)
	}

	if !contains(request.Smoker, constants.SmokerYes, constants.SmokerNo) {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuoteSmokerInvalid)
	}

	if !contains(request.OccupationClass, constants.OccupationLow, constants.OccupationStandard, constants.OccupationHigh) {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuoteOccupationInvalid)
	}

	if !contains(request.HealthRisk, constants.HealthRiskLow, constants.HealthRiskMedium, constants.HealthRiskHigh) {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuoteHealthRiskInvalid)
	}

	return request, nil
}

func contains(value string, allowedValues ...string) bool {
	if value == "" {
		return false
	}

	values := make([]interface{}, len(allowedValues))
	for index, allowedValue := range allowedValues {
		values[index] = allowedValue
	}

	return validation.Validate(value, validation.In(values...)) == nil
}
