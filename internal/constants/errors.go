package constants

import "errors"

const (
	ErrProductSlugRequired     = "product slug is required"
	ErrProductCategoryInvalid  = "category must be one of: life, health, vehicle"
	ErrProductFeaturedInvalid  = "featured must be true or false"
	ErrProductLimitInvalid     = "limit must be a positive integer"
	ErrProductLimitTooHigh     = "limit must be less than or equal to 50"
	ErrProductNotFound         = "product not found"
	ErrProductListFailed       = "failed to list products"
	ErrProductDetailFailed     = "failed to get product"
	ErrProductQuoteFailed      = "failed to create product quote"
	ErrProductQuoteBodyInvalid = "invalid product quote request body"
)

const (
	ErrQuoteAgeInvalid              = "age must be between 18 and 60"
	ErrQuoteGenderInvalid           = "gender must be male or female"
	ErrQuoteSmokerInvalid           = "smoker must be yes or no"
	ErrQuoteOccupationInvalid       = "occupation_class must be one of: low, standard, high"
	ErrQuoteHealthRiskInvalid       = "health_risk must be one of: low, medium, high"
	ErrQuotePaymentFrequencyInvalid = "payment_frequency must be one of: annual, semi_annual, quarterly, monthly"
	ErrQuoteSumAssuredInvalid       = "sum_assured must be greater than zero"
	ErrQuotePaymentTermInvalid      = "payment_term must be greater than zero"
	ErrQuoteSumAssuredOutOfRange    = "sum_assured is outside product allowed range"
	ErrQuotePaymentTermOutOfRange   = "payment_term is outside product allowed range"
	ErrQuotePricingRulesInvalid     = "product pricing rules are incomplete"
)

var (
	ErrApplicationStatusTransitionInvalidError     = errors.New(ErrApplicationStatusTransitionInvalid)
	ErrApplicationRejectionReasonRequiredError     = errors.New(ErrApplicationRejectionReasonRequired)
	ErrApplicationApprovalChecklistIncompleteError = errors.New(ErrApplicationApprovalChecklistIncomplete)
	ErrApplicationReviewCheckInvalidError          = errors.New(ErrApplicationReviewCheckInvalid)
	QuoteSumAssuredOutOfRangeError                 = errors.New(ErrQuoteSumAssuredOutOfRange)
	QuotePaymentTermOutOfRangeError                = errors.New(ErrQuotePaymentTermOutOfRange)
	QuotePricingRulesInvalidError                  = errors.New(ErrQuotePricingRulesInvalid)
)

func IsQuoteValidationError(err error) bool {
	return errors.Is(err, QuoteSumAssuredOutOfRangeError) ||
		errors.Is(err, QuotePaymentTermOutOfRangeError)
}

const (
	ErrApplicationBodyInvalid        = "invalid application request body"
	ErrApplicationRequired           = "application fields are invalid"
	ErrApplicationFullNameInvalid    = "full_name is invalid"
	ErrApplicationEmailInvalid       = "email is invalid"
	ErrApplicationPhoneInvalid       = "phone is invalid"
	ErrApplicationServiceUnavailable = "application service is unavailable"
	ErrApplicationCreateFailed       = "failed to create application"
	ErrApplicationNotFound           = "application not found"
	ErrApplicationGetFailed          = "failed to get application"
	ErrApplicationListFailed         = "failed to list applications"
	ErrApplicationListFilterInvalid  = "application list filter is invalid"
	ErrApplicationListLimitTooHigh   = "application list limit must be less than or equal to 50"
	ErrApplicationPageInvalid        = "page must be a positive integer"
	ErrApplicationLimitInvalid       = "limit must be a positive integer"
)

const (
	ErrApplicationStatusInvalid               = "application status is invalid"
	ErrApplicationStatusTransitionInvalid     = "application status transition is invalid"
	ErrApplicationRejectionReasonRequired     = "rejection_reason is required when rejecting"
	ErrApplicationStatusUpdateFailed          = "failed to update application status"
	ErrApplicationReviewCheckNotFound         = "application review check not found"
	ErrApplicationReviewCheckInvalid          = "application review check is invalid"
	ErrApplicationReviewCheckUpdateFailed     = "failed to update application review check"
	ErrApplicationApprovalChecklistIncomplete = "application review checklist is incomplete"
	ErrStorageServiceUnavailable              = "storage service unavailable"
	ErrStorageObjectNameRequired              = "object_name is required"
	ErrStoragePresignFailed                   = "failed to generate storage presigned url"
)
