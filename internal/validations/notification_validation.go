package validations

import (
	"errors"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

var validNotificationCategories = map[string]models.NotificationCategory{
	"underwriting": models.NotificationCategoryUnderwriting,
	"system":       models.NotificationCategorySystem,
	"knowledge":    models.NotificationCategoryKnowledge,
	"policy":       models.NotificationCategoryPolicy,
}

var validNotificationSeverities = map[string]models.NotificationSeverity{
	"INFO":     models.NotificationSeverityInfo,
	"WARNING":  models.NotificationSeverityWarning,
	"CRITICAL": models.NotificationSeverityCritical,
	"SUCCESS":  models.NotificationSeveritySuccess,
}

func ValidateCreateNotificationRequest(req *dtos.CreateNotificationRequest) error {
	if req == nil {
		return errors.New(constants.ErrNotificationTitleRequired)
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return errors.New(constants.ErrNotificationTitleRequired)
	}

	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" {
		return errors.New(constants.ErrNotificationTypeRequired)
	}

	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		return errors.New(constants.ErrNotificationMessageRequired)
	}

	req.Category = strings.ToLower(strings.TrimSpace(req.Category))
	if _, ok := validNotificationCategories[req.Category]; !ok {
		return errors.New(constants.ErrNotificationCategoryInvalid)
	}

	req.Severity = strings.ToUpper(strings.TrimSpace(req.Severity))
	if req.Severity == "" {
		req.Severity = string(models.NotificationSeverityInfo)
	} else if _, ok := validNotificationSeverities[req.Severity]; !ok {
		return errors.New(constants.ErrNotificationSeverityInvalid)
	}

	req.Link = strings.TrimSpace(req.Link)

	return nil
}

func ValidateNotificationQuery(query *dtos.NotificationQuery) error {
	if query == nil {
		return nil
	}

	query.Category = strings.ToLower(strings.TrimSpace(query.Category))
	if query.Category != "" && query.Category != "all" {
		if _, ok := validNotificationCategories[query.Category]; !ok {
			return errors.New(constants.ErrNotificationCategoryInvalid)
		}
	}

	query.Severity = strings.ToUpper(strings.TrimSpace(query.Severity))
	if query.Severity != "" && query.Severity != "ALL" {
		if _, ok := validNotificationSeverities[query.Severity]; !ok {
			return errors.New(constants.ErrNotificationSeverityInvalid)
		}
	}

	if query.Limit <= 0 {
		query.Limit = 20
	} else if query.Limit > 100 {
		return errors.New(constants.ErrNotificationLimitInvalid)
	}

	if query.Offset < 0 {
		query.Offset = 0
	}

	return nil
}
