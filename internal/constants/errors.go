package constants

const (
	ErrProductSlugRequired    = "product slug is required"
	ErrProductCategoryInvalid = "category must be one of: life, health, vehicle"
	ErrProductFeaturedInvalid = "featured must be true or false"
	ErrProductLimitInvalid    = "limit must be a positive integer"
	ErrProductLimitTooHigh    = "limit must be less than or equal to 50"
	ErrProductListFailed     = "failed to list products"
	ErrProductDetailFailed   = "failed to get product"
)
