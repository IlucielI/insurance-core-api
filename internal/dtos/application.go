package dtos

import "github.com/bayuanugerah/insurance-core-api/internal/models"

type CreateApplicationRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	ProductQuoteRequest
}

type ApplicationResponse struct {
	Data models.Application `json:"data"`
}

func ProductQuoteRequestToInput(request ProductQuoteRequest) CreateProductQuoteInput {
	return CreateProductQuoteInput{
		Age:              request.Age,
		Gender:           request.Gender,
		SumAssured:       request.SumAssured,
		PaymentTerm:      request.PaymentTerm,
		PaymentFrequency: request.PaymentFrequency,
		Smoker:           request.Smoker,
		OccupationClass:  request.OccupationClass,
		HealthRisk:       request.HealthRisk,
	}
}


type UpdateApplicationStatusRequest struct {
	Status          models.ApplicationStatus `json:"status"`
	ReviewedBy      string                   `json:"reviewed_by"`
	RejectionReason string                   `json:"rejection_reason"`
}
