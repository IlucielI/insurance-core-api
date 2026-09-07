package dtos

import "github.com/bayuanugerah/insurance-core-api/internal/models"

type QuestionnaireStepGroup struct {
	StepNumber  int               `json:"step_number"`
	PillarType  string            `json:"pillar_type"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Questions   []models.Question `json:"questions"`
}

type QuestionnaireResponse struct {
	ID          string                   `json:"id"`
	ProductID   *string                  `json:"product_id,omitempty"`
	ProductSlug string                   `json:"product_slug,omitempty"`
	Category    string                   `json:"category"`
	Title       string                   `json:"title"`
	Description string                   `json:"description,omitempty"`
	Version     int                      `json:"version"`
	Steps       []QuestionnaireStepGroup `json:"steps"`
	Questions   []models.Question        `json:"questions"`
}

type ApplicationAnswerInput struct {
	QuestionID string `json:"question_id"`
	Code       string `json:"code"`
	Value      any    `json:"value"`
}
