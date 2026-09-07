package dtos

import "github.com/bayuanugerah/insurance-core-api/internal/models"

type CreateApplicationRequest struct {
	ProductSlug string                   `json:"product_slug,omitempty"`
	ProductID   string                   `json:"product_id,omitempty"`
	FullName    string                   `json:"full_name"`
	Email       string                   `json:"email"`
	Phone       string                   `json:"phone"`
	Answers     []ApplicationAnswerInput `json:"answers,omitempty"`
	ProductQuoteRequest
}

type ApplicationResponse struct {
	Data models.Application `json:"data"`
}

func ProductQuoteRequestToInput(request ProductQuoteRequest) CreateProductQuoteInput {
	answers := make([]QuoteAnswerInput, len(request.Answers))
	copy(answers, request.Answers)

	hasRule := func(code string) bool {
		for _, a := range answers {
			if a.RuleCode == code {
				return true
			}
		}
		return false
	}

	if request.Smoker != "" && !hasRule("smoker") {
		answers = append(answers, QuoteAnswerInput{RuleCode: "smoker", Value: request.Smoker})
	}
	if request.OccupationClass != "" && !hasRule("occupation_class") {
		answers = append(answers, QuoteAnswerInput{RuleCode: "occupation_class", Value: request.OccupationClass})
	}
	if request.HealthRisk != "" && !hasRule("health_risk") {
		answers = append(answers, QuoteAnswerInput{RuleCode: "health_risk", Value: request.HealthRisk})
	}

	return CreateProductQuoteInput{
		Age:              request.Age,
		Gender:           request.Gender,
		SumAssured:       request.SumAssured,
		PaymentTerm:      request.PaymentTerm,
		PaymentFrequency: request.PaymentFrequency,
		Answers:          answers,
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
