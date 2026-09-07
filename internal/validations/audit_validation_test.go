package validations

import (
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCreateAuditLogRequest(t *testing.T) {
	t.Run("valid request with defaults", func(t *testing.T) {
		req := &dtos.CreateAuditLogRequest{
			ActorName:      "Budi Pratama",
			ActorRole:      "Senior Underwriter",
			Action:         "APPROVE_APPLICATION",
			Category:       "underwriting",
			TargetResource: "#APP-2026-8819",
		}
		err := ValidateCreateAuditLogRequest(req)
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1", req.IPAddress)
		assert.Equal(t, "SUCCESS", req.Status)
		assert.NotNil(t, req.Details)
	})

	t.Run("missing actor name", func(t *testing.T) {
		req := &dtos.CreateAuditLogRequest{
			ActorName:      " ",
			ActorRole:      "Underwriter",
			Action:         "APPROVE",
			Category:       "underwriting",
			TargetResource: "#APP-1",
		}
		err := ValidateCreateAuditLogRequest(req)
		require.Error(t, err)
		assert.Equal(t, constants.ErrAuditLogActorNameRequired, err.Error())
	})

	t.Run("missing actor role", func(t *testing.T) {
		req := &dtos.CreateAuditLogRequest{
			ActorName:      "Budi",
			ActorRole:      "",
			Action:         "APPROVE",
			Category:       "underwriting",
			TargetResource: "#APP-1",
		}
		err := ValidateCreateAuditLogRequest(req)
		require.Error(t, err)
		assert.Equal(t, constants.ErrAuditLogActorRoleRequired, err.Error())
	})

	t.Run("missing action", func(t *testing.T) {
		req := &dtos.CreateAuditLogRequest{
			ActorName:      "Budi",
			ActorRole:      "Underwriter",
			Action:         "",
			Category:       "underwriting",
			TargetResource: "#APP-1",
		}
		err := ValidateCreateAuditLogRequest(req)
		require.Error(t, err)
		assert.Equal(t, constants.ErrAuditLogActionRequired, err.Error())
	})

	t.Run("invalid category", func(t *testing.T) {
		req := &dtos.CreateAuditLogRequest{
			ActorName:      "Budi",
			ActorRole:      "Underwriter",
			Action:         "APPROVE",
			Category:       "invalid_cat",
			TargetResource: "#APP-1",
		}
		err := ValidateCreateAuditLogRequest(req)
		require.Error(t, err)
		assert.Equal(t, constants.ErrAuditLogCategoryInvalid, err.Error())
	})

	t.Run("missing target resource", func(t *testing.T) {
		req := &dtos.CreateAuditLogRequest{
			ActorName:      "Budi",
			ActorRole:      "Underwriter",
			Action:         "APPROVE",
			Category:       "underwriting",
			TargetResource: "",
		}
		err := ValidateCreateAuditLogRequest(req)
		require.Error(t, err)
		assert.Equal(t, constants.ErrAuditLogTargetRequired, err.Error())
	})

	t.Run("invalid status", func(t *testing.T) {
		req := &dtos.CreateAuditLogRequest{
			ActorName:      "Budi",
			ActorRole:      "Underwriter",
			Action:         "APPROVE",
			Category:       "underwriting",
			TargetResource: "#APP-1",
			Status:         "UNKNOWN",
		}
		err := ValidateCreateAuditLogRequest(req)
		require.Error(t, err)
		assert.Equal(t, constants.ErrAuditLogStatusInvalid, err.Error())
	})

	t.Run("nil request", func(t *testing.T) {
		err := ValidateCreateAuditLogRequest(nil)
		require.Error(t, err)
	})
}

func TestValidateAuditLogQuery(t *testing.T) {
	t.Run("nil query passes", func(t *testing.T) {
		err := ValidateAuditLogQuery(nil)
		assert.NoError(t, err)
	})

	t.Run("default limit and offset", func(t *testing.T) {
		q := &dtos.AuditLogQuery{
			Limit:  0,
			Offset: -5,
		}
		err := ValidateAuditLogQuery(q)
		require.NoError(t, err)
		assert.Equal(t, 20, q.Limit)
		assert.Equal(t, 0, q.Offset)
	})

	t.Run("limit too high", func(t *testing.T) {
		q := &dtos.AuditLogQuery{
			Limit: 150,
		}
		err := ValidateAuditLogQuery(q)
		require.Error(t, err)
		assert.Equal(t, constants.ErrAuditLogQueryLimitInvalid, err.Error())
	})

	t.Run("invalid category filter", func(t *testing.T) {
		q := &dtos.AuditLogQuery{
			Category: "unknown_cat",
		}
		err := ValidateAuditLogQuery(q)
		require.Error(t, err)
		assert.Equal(t, constants.ErrAuditLogCategoryInvalid, err.Error())
	})

	t.Run("all category filter allowed", func(t *testing.T) {
		q := &dtos.AuditLogQuery{
			Category: "all",
			Status:   "all",
		}
		err := ValidateAuditLogQuery(q)
		assert.NoError(t, err)
	})

	t.Run("invalid status filter", func(t *testing.T) {
		q := &dtos.AuditLogQuery{
			Status: "NOT_A_STATUS",
		}
		err := ValidateAuditLogQuery(q)
		require.Error(t, err)
		assert.Equal(t, constants.ErrAuditLogStatusInvalid, err.Error())
	})
}

func TestCalculateAuditHashAndJSON(t *testing.T) {
	ts := time.Date(2026, 9, 6, 6, 55, 12, 0, time.UTC)
	hash1 := CalculateAuditHash(ts, "Budi", "Senior Underwriter", "APPROVE", "underwriting", "#APP-1", "SUCCESS", "{}")
	assert.Len(t, hash1, 64)

	// Same input produces exact same hash
	hash2 := CalculateAuditHash(ts, "Budi", "Senior Underwriter", "APPROVE", "underwriting", "#APP-1", "SUCCESS", "{}")
	assert.Equal(t, hash1, hash2)

	// Modified input produces different hash
	hash3 := CalculateAuditHash(ts, "Budi Santoso", "Senior Underwriter", "APPROVE", "underwriting", "#APP-1", "SUCCESS", "{}")
	assert.NotEqual(t, hash1, hash3)

	jsonStr, err := DetailsToJSON(map[string]any{"key": "val"})
	require.NoError(t, err)
	assert.Contains(t, jsonStr, `"key":"val"`)

	emptyJSON, err := DetailsToJSON(nil)
	require.NoError(t, err)
	assert.Equal(t, "{}", emptyJSON)
}
