package dtos

import "github.com/bayuanugerah/insurance-core-api/internal/models"

type UpdateApplicationReviewCheckRequest struct {
	Status     models.ApplicationReviewCheckStatus `json:"status"`
	Notes      string                              `json:"notes"`
	ReviewedBy string                              `json:"reviewed_by"`
}
