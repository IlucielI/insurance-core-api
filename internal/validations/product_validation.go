package validations

import (
	"errors"
	"strconv"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

type ProductListQuery struct {
	Category   string
	IsFeatured *bool
	Limit      int
}

func ValidateProductListQuery(categoryValue string, featuredValue string, limitValue string) (ProductListQuery, error) {
	query := ProductListQuery{
		Category: strings.TrimSpace(categoryValue),
	}

	if query.Category != "" && !isValidProductCategory(query.Category) {
		return ProductListQuery{}, errors.New(constants.ErrProductCategoryInvalid)
	}

	featuredValue = strings.TrimSpace(featuredValue)
	if featuredValue != "" {
		isFeatured, err := strconv.ParseBool(featuredValue)
		if err != nil {
			return ProductListQuery{}, errors.New(constants.ErrProductFeaturedInvalid)
		}

		query.IsFeatured = &isFeatured
	}

	limitValue = strings.TrimSpace(limitValue)
	if limitValue != "" {
		limit, err := strconv.Atoi(limitValue)
		if err != nil || limit < 1 {
			return ProductListQuery{}, errors.New(constants.ErrProductLimitInvalid)
		}

		if limit > constants.MaxProductListLimit {
			return ProductListQuery{}, errors.New(constants.ErrProductLimitTooHigh)
		}

		query.Limit = limit
	}

	return query, nil
}

func ValidateProductSlug(slug string) (string, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", errors.New(constants.ErrProductSlugRequired)
	}

	return slug, nil
}

func isValidProductCategory(category string) bool {
	switch models.ProductCategory(category) {
	case models.ProductCategoryLife, models.ProductCategoryHealth, models.ProductCategoryVehicle:
		return true
	default:
		return false
	}
}
