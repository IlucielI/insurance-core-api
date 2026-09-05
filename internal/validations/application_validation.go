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

func ValidateApplicationListQuery(statusValue, productIDValue, pageValue, limitValue string) (dtos.ApplicationListQuery, error) {
	query := dtos.ApplicationListQuery{
		Status:    strings.TrimSpace(statusValue),
		ProductID: strings.TrimSpace(productIDValue),
	}

	if query.Status != "" && !validApplicationStatusFilter(models.ApplicationStatus(query.Status)) {
		return dtos.ApplicationListQuery{}, errors.New(constants.ErrApplicationStatusInvalid)
	}

	if query.ProductID != "" && validation.Validate(query.ProductID, validation.Length(1, 64)) != nil {
		return dtos.ApplicationListQuery{}, errors.New(constants.ErrApplicationListFilterInvalid)
	}

	page, err := parsePositiveInt(pageValue, constants.ErrApplicationPageInvalid)
	if err != nil {
		return dtos.ApplicationListQuery{}, err
	}
	if page == 0 {
		page = 1
	}
	limit, err := parsePositiveInt(limitValue, constants.ErrApplicationLimitInvalid)
	if err != nil {
		return dtos.ApplicationListQuery{}, err
	}
	if limit == 0 {
		limit = constants.DefaultApplicationListLimit
	}
	if limit > constants.MaxApplicationListLimit {
		return dtos.ApplicationListQuery{}, errors.New(constants.ErrApplicationListLimitTooHigh)
	}

	query.Page = page
	query.Limit = limit
	return query, nil
}

func parsePositiveInt(value string, errorMessage string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, errors.New(errorMessage)
	}

	return parsed, nil
}

func validApplicationStatusFilter(status models.ApplicationStatus) bool {
	switch status {
	case models.ApplicationStatusSubmitted,
		models.ApplicationStatusUnderReview,
		models.ApplicationStatusApproved,
		models.ApplicationStatusRejected:
		return true
	default:
		return false
	}
}
