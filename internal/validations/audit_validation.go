package validations

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

var validAuditCategories = map[string]models.AuditCategory{
	"underwriting": models.AuditCategoryUnderwriting,
	"product":      models.AuditCategoryProduct,
	"knowledge":    models.AuditCategoryKnowledge,
	"auth":         models.AuditCategoryAuth,
	"system":       models.AuditCategorySystem,
}

var validAuditStatuses = map[string]models.AuditSeverity{
	"SUCCESS": models.AuditSeveritySuccess,
	"WARNING": models.AuditSeverityWarning,
	"FAILED":  models.AuditSeverityFailed,
}

func ValidateCreateAuditLogRequest(req *dtos.CreateAuditLogRequest) error {
	if req == nil {
		return errors.New(constants.ErrAuditLogActorNameRequired)
	}

	req.ActorName = strings.TrimSpace(req.ActorName)
	if req.ActorName == "" {
		return errors.New(constants.ErrAuditLogActorNameRequired)
	}

	req.ActorRole = strings.TrimSpace(req.ActorRole)
	if req.ActorRole == "" {
		return errors.New(constants.ErrAuditLogActorRoleRequired)
	}

	req.Action = strings.TrimSpace(req.Action)
	if req.Action == "" {
		return errors.New(constants.ErrAuditLogActionRequired)
	}

	req.Category = strings.ToLower(strings.TrimSpace(req.Category))
	if _, ok := validAuditCategories[req.Category]; !ok {
		return errors.New(constants.ErrAuditLogCategoryInvalid)
	}

	req.TargetResource = strings.TrimSpace(req.TargetResource)
	if req.TargetResource == "" {
		return errors.New(constants.ErrAuditLogTargetRequired)
	}

	req.IPAddress = strings.TrimSpace(req.IPAddress)
	if req.IPAddress == "" {
		req.IPAddress = "127.0.0.1"
	}

	req.Status = strings.ToUpper(strings.TrimSpace(req.Status))
	if req.Status == "" {
		req.Status = string(models.AuditSeveritySuccess)
	} else if _, ok := validAuditStatuses[req.Status]; !ok {
		return errors.New(constants.ErrAuditLogStatusInvalid)
	}

	if req.Details == nil {
		req.Details = make(map[string]any)
	}

	return nil
}

func ValidateAuditLogQuery(query *dtos.AuditLogQuery) error {
	if query == nil {
		return nil
	}

	query.Category = strings.ToLower(strings.TrimSpace(query.Category))
	if query.Category != "" && query.Category != "all" {
		if _, ok := validAuditCategories[query.Category]; !ok {
			return errors.New(constants.ErrAuditLogCategoryInvalid)
		}
	}

	query.Status = strings.ToUpper(strings.TrimSpace(query.Status))
	if query.Status != "" && query.Status != "ALL" {
		if _, ok := validAuditStatuses[query.Status]; !ok {
			return errors.New(constants.ErrAuditLogStatusInvalid)
		}
	}

	query.Search = strings.TrimSpace(query.Search)

	if query.Limit <= 0 {
		query.Limit = 20
	} else if query.Limit > 100 {
		return errors.New(constants.ErrAuditLogQueryLimitInvalid)
	}

	if query.Offset < 0 {
		query.Offset = 0
	}

	return nil
}

func CalculateAuditHash(timestamp time.Time, actorName, actorRole, action, category, target, status, detailsJSON string) string {
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s",
		timestamp.UTC().Format(time.RFC3339Nano),
		actorName,
		actorRole,
		action,
		category,
		target,
		status,
		detailsJSON,
	)
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}

func DetailsToJSON(details map[string]any) (string, error) {
	if len(details) == 0 {
		return "{}", nil
	}
	bytes, err := json.Marshal(details)
	if err != nil {
		return "{}", err
	}
	return string(bytes), nil
}
