package dtos

import "github.com/bayuanugerah/insurance-core-api/internal/models"

type ApplicationListQuery struct {
	Status    string
	ProductID string
	Page      int
	Limit     int
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ApplicationListResponse struct {
	Data []models.Application `json:"data"`
	Meta PaginationMeta       `json:"meta"`
}
