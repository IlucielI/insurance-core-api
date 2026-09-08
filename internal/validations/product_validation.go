package validations

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidateProductListQuery(categoryValue string, featuredValue string, limitValue string, searchValue string, statusValue ...string) (dtos.ProductListQuery, error) {
	query := dtos.ProductListQuery{
		Category: strings.TrimSpace(categoryValue),
	}

	if query.Category != "" && validation.Validate(query.Category, validation.In(
		string(models.ProductCategoryLife),
		string(models.ProductCategoryHealth),
		string(models.ProductCategoryVehicle),
	)) != nil {
		return dtos.ProductListQuery{}, errors.New(constants.ErrProductCategoryInvalid)
	}

	query.Status = string(models.ProductStatusActive)
	if len(statusValue) > 0 {
		status := strings.ToLower(strings.TrimSpace(statusValue[0]))
		if status != "" && status != "all" {
			if validation.Validate(status, validation.In(
				string(models.ProductStatusActive),
				string(models.ProductStatusDraft),
				string(models.ProductStatusArchived),
			)) != nil {
				return dtos.ProductListQuery{}, errors.New(constants.ErrProductStatusInvalid)
			}
			query.Status = status
		} else if status == "all" {
			query.Status = "all"
		}
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

	searchValue = strings.TrimSpace(searchValue)
	if len(searchValue) > constants.MaxProductSearchLength {
		return dtos.ProductListQuery{}, errors.New(constants.ErrProductSearchInvalid)
	}
	query.Search = searchValue

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
	request.Gender = strings.ToLower(strings.TrimSpace(request.Gender))
	request.PaymentFrequency = strings.ToLower(strings.TrimSpace(request.PaymentFrequency))
	request.Smoker = strings.ToLower(strings.TrimSpace(request.Smoker))
	request.OccupationClass = strings.ToLower(strings.TrimSpace(request.OccupationClass))
	request.HealthRisk = strings.ToLower(strings.TrimSpace(request.HealthRisk))

	if request.PaymentFrequency == "annually" {
		request.PaymentFrequency = constants.PaymentFrequencyAnnual
	}

	if request.OccupationClass == "" {
		request.OccupationClass = constants.OccupationStandard
	}

	if request.HealthRisk == "" {
		request.HealthRisk = constants.HealthRiskLow
	}

	if request.Smoker == "" {
		request.Smoker = constants.SmokerNo
	}

	if request.Gender == "" {
		request.Gender = constants.GenderMale
	}

	if request.Age < 0 {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuoteAgeInvalid)
	}

	if !contains(request.Gender, constants.GenderMale, constants.GenderFemale) {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuoteGenderInvalid)
	}

	if validation.Validate(request.SumAssured, validation.Required, validation.Min(int64(1))) != nil {
		return dtos.ProductQuoteRequest{}, errors.New(constants.ErrQuoteSumAssuredInvalid)
	}

	if validation.Validate(request.PaymentTerm, validation.Required, validation.Min(1)) != nil {
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

func ValidateApplicationRequest(request dtos.CreateApplicationRequest) (dtos.CreateApplicationRequest, error) {
	request.ProductSlug = strings.TrimSpace(request.ProductSlug)
	request.ProductID = strings.TrimSpace(request.ProductID)
	request.FullName = strings.TrimSpace(request.FullName)
	request.Email = strings.TrimSpace(request.Email)
	request.Phone = strings.TrimSpace(request.Phone)
	if validation.Validate(request.FullName, validation.Required, validation.Length(2, 120)) != nil {
		return dtos.CreateApplicationRequest{}, errors.New(constants.ErrApplicationFullNameInvalid)
	}
	if validation.Validate(request.Email, validation.Required, validation.Length(3, 255), is.EmailFormat) != nil {
		return dtos.CreateApplicationRequest{}, errors.New(constants.ErrApplicationEmailInvalid)
	}
	if validation.Validate(request.Phone, validation.Required, validation.Length(7, 32)) != nil {
		return dtos.CreateApplicationRequest{}, errors.New(constants.ErrApplicationPhoneInvalid)
	}

	quote, err := ValidateProductQuoteRequest(request.ProductQuoteRequest)
	if err != nil {
		return dtos.CreateApplicationRequest{}, err
	}
	request.ProductQuoteRequest = quote
	return request, nil
}

func ValidateApplicationReviewCheckRequest(request dtos.UpdateApplicationReviewCheckRequest) (dtos.UpdateApplicationReviewCheckRequest, error) {
	request.Status = models.ApplicationReviewCheckStatus(strings.TrimSpace(string(request.Status)))
	request.Notes = strings.TrimSpace(request.Notes)
	request.ReviewedBy = strings.TrimSpace(request.ReviewedBy)

	if !validApplicationReviewCheckStatus(request.Status) || validation.Validate(request.ReviewedBy, validation.Required, validation.Length(2, 120)) != nil {
		return dtos.UpdateApplicationReviewCheckRequest{}, constants.ErrApplicationReviewCheckInvalidError
	}

	if validation.Validate(request.Notes, validation.Length(0, 500)) != nil {
		return dtos.UpdateApplicationReviewCheckRequest{}, constants.ErrApplicationReviewCheckInvalidError
	}

	return request, nil
}

func ValidateApplicationReviewCheckType(value string) (models.ApplicationReviewCheckType, error) {
	checkType := models.ApplicationReviewCheckType(strings.TrimSpace(value))
	if !validApplicationReviewCheckType(checkType) {
		return "", constants.ErrApplicationReviewCheckInvalidError
	}

	return checkType, nil
}

func validApplicationReviewCheckType(checkType models.ApplicationReviewCheckType) bool {
	switch checkType {
	case models.ApplicationReviewCheckTypeIdentityVerified,
		models.ApplicationReviewCheckTypeIncomeVerified,
		models.ApplicationReviewCheckTypeDocumentsComplete,
		models.ApplicationReviewCheckTypeMedicalRequired:
		return true
	default:
		return false
	}
}

func validApplicationReviewCheckStatus(status models.ApplicationReviewCheckStatus) bool {
	switch status {
	case models.ApplicationReviewCheckStatusPending,
		models.ApplicationReviewCheckStatusPassed,
		models.ApplicationReviewCheckStatusFailed,
		models.ApplicationReviewCheckStatusNotNeeded:
		return true
	default:
		return false
	}
}

func ValidateProductStatus(status string) (models.ProductStatus, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if validation.Validate(status, validation.Required, validation.In(
		string(models.ProductStatusActive),
		string(models.ProductStatusDraft),
		string(models.ProductStatusArchived),
	)) != nil {
		return "", errors.New(constants.ErrProductStatusInvalid)
	}
	return models.ProductStatus(status), nil
}

func ValidateCreateProductRequest(req dtos.CreateProductRequest) (dtos.CreateProductRequest, error) {
	req.Name = strings.TrimSpace(req.Name)
	if validation.Validate(req.Name, validation.Required, validation.Length(3, 120)) != nil {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductNameInvalid)
	}

	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	if req.Slug == "" || !slugRegex.MatchString(req.Slug) || len(req.Slug) > 120 {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductSlugInvalid)
	}

	req.Category = strings.ToLower(strings.TrimSpace(req.Category))
	if validation.Validate(req.Category, validation.Required, validation.In(
		string(models.ProductCategoryLife),
		string(models.ProductCategoryHealth),
		string(models.ProductCategoryVehicle),
	)) != nil {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductCategoryInvalid)
	}

	if req.Status == "" {
		req.Status = string(models.ProductStatusActive)
	} else {
		status, err := ValidateProductStatus(req.Status)
		if err != nil {
			return dtos.CreateProductRequest{}, err
		}
		req.Status = string(status)
	}

	req.ShortDescription = strings.TrimSpace(req.ShortDescription)
	if validation.Validate(req.ShortDescription, validation.Required, validation.Length(3, 255)) != nil {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductShortDescriptionRequired)
	}

	req.Description = strings.TrimSpace(req.Description)
	if validation.Validate(req.Description, validation.Required) != nil {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductDescriptionRequired)
	}

	req.TargetCustomer = strings.TrimSpace(req.TargetCustomer)
	if validation.Validate(req.TargetCustomer, validation.Required, validation.Length(3, 255)) != nil {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductTargetCustomerRequired)
	}

	if req.MinSumAssured <= 0 {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductMinSumAssuredInvalid)
	}
	if req.MaxSumAssured < req.MinSumAssured {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductMaxSumAssuredInvalid)
	}

	if req.MinPaymentTerm < 1 {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductMinPaymentTermInvalid)
	}
	if req.MaxPaymentTerm < req.MinPaymentTerm {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductMaxPaymentTermInvalid)
	}

	if req.StartingPremium <= 0 {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductStartingPremiumInvalid)
	}

	if req.NonMCULimit < 0 {
		return dtos.CreateProductRequest{}, errors.New(constants.ErrProductNonMCULimitInvalid)
	}
	if req.NonMCULimit == 0 {
		req.NonMCULimit = 500000000
	}

	cleanedBenefits := make([]string, 0, len(req.Benefits))
	for _, b := range req.Benefits {
		b = strings.TrimSpace(b)
		if b != "" {
			cleanedBenefits = append(cleanedBenefits, b)
		}
	}
	req.Benefits = cleanedBenefits

	cleanedExclusions := make([]string, 0, len(req.Exclusions))
	for _, e := range req.Exclusions {
		e = strings.TrimSpace(e)
		if e != "" {
			cleanedExclusions = append(cleanedExclusions, e)
		}
	}
	req.Exclusions = cleanedExclusions

	return req, nil
}

func ValidateUpdateProductRequest(req dtos.UpdateProductRequest) (dtos.UpdateProductRequest, error) {
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if validation.Validate(name, validation.Required, validation.Length(3, 120)) != nil {
			return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductNameInvalid)
		}
		req.Name = &name
	}

	if req.Slug != nil {
		slug := strings.ToLower(strings.TrimSpace(*req.Slug))
		if slug == "" || !slugRegex.MatchString(slug) || len(slug) > 120 {
			return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductSlugInvalid)
		}
		req.Slug = &slug
	}

	if req.Category != nil {
		category := strings.ToLower(strings.TrimSpace(*req.Category))
		if validation.Validate(category, validation.Required, validation.In(
			string(models.ProductCategoryLife),
			string(models.ProductCategoryHealth),
			string(models.ProductCategoryVehicle),
		)) != nil {
			return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductCategoryInvalid)
		}
		req.Category = &category
	}

	if req.Status != nil {
		status, err := ValidateProductStatus(*req.Status)
		if err != nil {
			return dtos.UpdateProductRequest{}, err
		}
		statusStr := string(status)
		req.Status = &statusStr
	}

	if req.ShortDescription != nil {
		sd := strings.TrimSpace(*req.ShortDescription)
		if validation.Validate(sd, validation.Required, validation.Length(3, 255)) != nil {
			return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductShortDescriptionRequired)
		}
		req.ShortDescription = &sd
	}

	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		if validation.Validate(desc, validation.Required) != nil {
			return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductDescriptionRequired)
		}
		req.Description = &desc
	}

	if req.TargetCustomer != nil {
		tc := strings.TrimSpace(*req.TargetCustomer)
		if validation.Validate(tc, validation.Required, validation.Length(3, 255)) != nil {
			return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductTargetCustomerRequired)
		}
		req.TargetCustomer = &tc
	}

	if req.MinSumAssured != nil && *req.MinSumAssured <= 0 {
		return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductMinSumAssuredInvalid)
	}

	if req.MaxSumAssured != nil && *req.MaxSumAssured <= 0 {
		return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductMaxSumAssuredInvalid)
	}

	if req.MinSumAssured != nil && req.MaxSumAssured != nil && *req.MaxSumAssured < *req.MinSumAssured {
		return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductMaxSumAssuredInvalid)
	}

	if req.MinPaymentTerm != nil && *req.MinPaymentTerm < 1 {
		return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductMinPaymentTermInvalid)
	}

	if req.MaxPaymentTerm != nil && *req.MaxPaymentTerm < 1 {
		return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductMaxPaymentTermInvalid)
	}

	if req.MinPaymentTerm != nil && req.MaxPaymentTerm != nil && *req.MaxPaymentTerm < *req.MinPaymentTerm {
		return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductMaxPaymentTermInvalid)
	}

	if req.StartingPremium != nil && *req.StartingPremium <= 0 {
		return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductStartingPremiumInvalid)
	}

	if req.NonMCULimit != nil && *req.NonMCULimit < 0 {
		return dtos.UpdateProductRequest{}, errors.New(constants.ErrProductNonMCULimitInvalid)
	}

	if req.Benefits != nil {
		cleaned := make([]string, 0, len(*req.Benefits))
		for _, b := range *req.Benefits {
			b = strings.TrimSpace(b)
			if b != "" {
				cleaned = append(cleaned, b)
			}
		}
		req.Benefits = &cleaned
	}

	if req.Exclusions != nil {
		cleaned := make([]string, 0, len(*req.Exclusions))
		for _, e := range *req.Exclusions {
			e = strings.TrimSpace(e)
			if e != "" {
				cleaned = append(cleaned, e)
			}
		}
		req.Exclusions = &cleaned
	}

	return req, nil
}
