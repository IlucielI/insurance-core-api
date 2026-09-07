package constants

import "errors"

const (
	ErrProductSlugRequired     = "product slug is required"
	ErrProductCategoryInvalid  = "category must be one of: life, health, vehicle"
	ErrProductFeaturedInvalid  = "featured must be true or false"
	ErrProductLimitInvalid     = "limit must be a positive integer"
	ErrProductLimitTooHigh     = "limit must be less than or equal to 50"
	ErrProductSearchInvalid    = "search must be 100 characters or less"
	ErrProductNotFound         = "product not found"
	ErrProductListFailed       = "failed to list products"
	ErrProductDetailFailed     = "failed to get product"
	ErrProductQuoteFailed      = "failed to create product quote"
	ErrProductQuoteBodyInvalid = "invalid product quote request body"
	ErrProductNameRequired     = "product name is required"
	ErrProductNameInvalid      = "product name must be between 3 and 120 characters"
	ErrProductSlugInvalid      = "product slug must be a valid lowercase kebab-case string"
	ErrProductSlugAlreadyExists = "product slug already exists"
	ErrProductStatusInvalid    = "status must be one of: active, draft, archived"
	ErrProductShortDescriptionRequired = "short_description is required"
	ErrProductDescriptionRequired      = "description is required"
	ErrProductTargetCustomerRequired   = "target_customer is required"
	ErrProductMinSumAssuredInvalid     = "min_sum_assured must be greater than zero"
	ErrProductMaxSumAssuredInvalid     = "max_sum_assured must be greater than or equal to min_sum_assured"
	ErrProductMinPaymentTermInvalid    = "min_payment_term must be at least 1"
	ErrProductMaxPaymentTermInvalid    = "max_payment_term must be greater than or equal to min_payment_term"
	ErrProductStartingPremiumInvalid  = "starting_premium must be greater than zero"
	ErrProductNonMCULimitInvalid       = "non_mcu_limit must be greater than or equal to zero"
	ErrProductHasApplications          = "cannot delete product with existing applications; archive it instead"
	ErrProductCreateFailed             = "failed to create product"
	ErrProductUpdateFailed             = "failed to update product"
	ErrProductDeleteFailed             = "failed to delete product"
	ErrProductMetricsFailed            = "failed to get product metrics"
	ErrProductPricingRulesUpdateFailed = "failed to update pricing rules"
	ErrProductIDRequired               = "product id is required"
	ErrProductRequestBodyInvalid       = "invalid product request body"
	ErrPricingRulesBodyInvalid         = "invalid pricing rules body"
)

const (
	ErrKnowledgeDocNotFound           = "knowledge document not found"
	ErrKnowledgeDocSlugAlreadyExists   = "knowledge document slug already exists"
	ErrKnowledgeDocTitleRequired      = "title is required"
	ErrKnowledgeDocTitleInvalid       = "title must be between 3 and 255 characters"
	ErrKnowledgeDocSlugInvalid        = "slug must be a valid lowercase kebab-case string"
	ErrKnowledgeDocCategoryInvalid    = "category must be one of: underwriting, product, claim_faq, compliance, company"
	ErrKnowledgeDocStatusInvalid      = "status must be one of: indexed, syncing, draft"
	ErrKnowledgeDocSummaryRequired    = "summary is required"
	ErrKnowledgeDocContentRequired    = "content is required"
	ErrKnowledgeDocIDRequired         = "document id is required"
	ErrKnowledgeDocRequestBodyInvalid = "invalid knowledge document request body"
	ErrKnowledgeDocListFailed         = "failed to list knowledge documents"
	ErrKnowledgeDocDetailFailed       = "failed to get knowledge document"
	ErrKnowledgeDocCreateFailed       = "failed to create knowledge document"
	ErrKnowledgeDocUpdateFailed       = "failed to update knowledge document"
	ErrKnowledgeDocDeleteFailed       = "failed to delete knowledge document"
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
	ErrProductSlugAlreadyExistsError               = errors.New(ErrProductSlugAlreadyExists)
	ErrProductHasApplicationsError                 = errors.New(ErrProductHasApplications)
	ErrProductNotFoundError                        = errors.New(ErrProductNotFound)
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

	ErrAuditLogNotFound          = "audit log not found"
	ErrAuditLogActorNameRequired = "actor_name is required"
	ErrAuditLogActorRoleRequired = "actor_role is required"
	ErrAuditLogActionRequired    = "action is required"
	ErrAuditLogCategoryInvalid   = "category must be one of: underwriting, product, knowledge, auth, system"
	ErrAuditLogStatusInvalid     = "status must be one of: SUCCESS, WARNING, FAILED"
	ErrAuditLogTargetRequired    = "target_resource is required"
	ErrAuditLogQueryLimitInvalid = "limit must be between 1 and 100"
)

var (
	ErrAuditLogNotFoundError = errors.New(ErrAuditLogNotFound)
)

