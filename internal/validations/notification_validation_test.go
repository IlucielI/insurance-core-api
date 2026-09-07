package validations

import (
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/stretchr/testify/assert"
)

func TestValidateCreateNotificationRequest(t *testing.T) {
	t.Run("nil request returns title required error", func(t *testing.T) {
		err := ValidateCreateNotificationRequest(nil)
		assert.EqualError(t, err, constants.ErrNotificationTitleRequired)
	})

	t.Run("empty title returns error", func(t *testing.T) {
		req := &dtos.CreateNotificationRequest{
			Title: "   ",
		}
		err := ValidateCreateNotificationRequest(req)
		assert.EqualError(t, err, constants.ErrNotificationTitleRequired)
	})

	t.Run("empty type returns error", func(t *testing.T) {
		req := &dtos.CreateNotificationRequest{
			Title: "Test Title",
			Type:  "",
		}
		err := ValidateCreateNotificationRequest(req)
		assert.EqualError(t, err, constants.ErrNotificationTypeRequired)
	})

	t.Run("empty message returns error", func(t *testing.T) {
		req := &dtos.CreateNotificationRequest{
			Title:   "Test Title",
			Type:    "APPLICATION_SUBMITTED",
			Message: "  ",
		}
		err := ValidateCreateNotificationRequest(req)
		assert.EqualError(t, err, constants.ErrNotificationMessageRequired)
	})

	t.Run("invalid category returns error", func(t *testing.T) {
		req := &dtos.CreateNotificationRequest{
			Title:    "Test Title",
			Type:     "APPLICATION_SUBMITTED",
			Message:  "Some message",
			Category: "unknown_category",
		}
		err := ValidateCreateNotificationRequest(req)
		assert.EqualError(t, err, constants.ErrNotificationCategoryInvalid)
	})

	t.Run("invalid severity returns error", func(t *testing.T) {
		req := &dtos.CreateNotificationRequest{
			Title:    "Test Title",
			Type:     "APPLICATION_SUBMITTED",
			Message:  "Some message",
			Category: "underwriting",
			Severity: "INVALID_SEVERITY",
		}
		err := ValidateCreateNotificationRequest(req)
		assert.EqualError(t, err, constants.ErrNotificationSeverityInvalid)
	})

	t.Run("valid request with default severity", func(t *testing.T) {
		req := &dtos.CreateNotificationRequest{
			Title:    "New Application",
			Type:     "APPLICATION_SUBMITTED",
			Message:  "Details here",
			Category: "underwriting",
		}
		err := ValidateCreateNotificationRequest(req)
		assert.NoError(t, err)
		assert.Equal(t, "INFO", req.Severity)
	})

	t.Run("valid request with explicit severity", func(t *testing.T) {
		req := &dtos.CreateNotificationRequest{
			Title:    "Warning alert",
			Type:     "SLA_WARNING",
			Message:  "Details here",
			Category: "system",
			Severity: "warning",
		}
		err := ValidateCreateNotificationRequest(req)
		assert.NoError(t, err)
		assert.Equal(t, "WARNING", req.Severity)
	})
}

func TestValidateNotificationQuery(t *testing.T) {
	t.Run("nil query succeeds", func(t *testing.T) {
		err := ValidateNotificationQuery(nil)
		assert.NoError(t, err)
	})

	t.Run("default limit and offset", func(t *testing.T) {
		q := &dtos.NotificationQuery{
			Limit:  0,
			Offset: -5,
		}
		err := ValidateNotificationQuery(q)
		assert.NoError(t, err)
		assert.Equal(t, 20, q.Limit)
		assert.Equal(t, 0, q.Offset)
	})

	t.Run("limit too high returns error", func(t *testing.T) {
		q := &dtos.NotificationQuery{
			Limit: 101,
		}
		err := ValidateNotificationQuery(q)
		assert.EqualError(t, err, constants.ErrNotificationLimitInvalid)
	})

	t.Run("invalid category returns error", func(t *testing.T) {
		q := &dtos.NotificationQuery{
			Category: "invalid_cat",
		}
		err := ValidateNotificationQuery(q)
		assert.EqualError(t, err, constants.ErrNotificationCategoryInvalid)
	})

	t.Run("invalid severity returns error", func(t *testing.T) {
		q := &dtos.NotificationQuery{
			Severity: "danger",
		}
		err := ValidateNotificationQuery(q)
		assert.EqualError(t, err, constants.ErrNotificationSeverityInvalid)
	})

	t.Run("valid query with all filter options", func(t *testing.T) {
		q := &dtos.NotificationQuery{
			Category: "all",
			Severity: "ALL",
			Limit:    50,
		}
		err := ValidateNotificationQuery(q)
		assert.NoError(t, err)
		assert.Equal(t, 50, q.Limit)
	})
}
