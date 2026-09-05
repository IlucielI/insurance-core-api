package dtos

import (
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
)

func TestProductQuoteRequestToInput(t *testing.T) {
	request := ProductQuoteRequest{
		Age:              35,
		Gender:           constants.GenderMale,
		SumAssured:       300000000,
		PaymentTerm:      10,
		PaymentFrequency: constants.PaymentFrequencyMonthly,
		Smoker:           constants.SmokerNo,
		OccupationClass:  constants.OccupationStandard,
		HealthRisk:       constants.HealthRiskLow,
	}

	input := ProductQuoteRequestToInput(request)
	if input.Age != request.Age || input.Gender != request.Gender || input.SumAssured != request.SumAssured || input.PaymentTerm != request.PaymentTerm || input.PaymentFrequency != request.PaymentFrequency || input.Smoker != request.Smoker || input.OccupationClass != request.OccupationClass || input.HealthRisk != request.HealthRisk {
		t.Fatalf("ProductQuoteRequestToInput() = %+v, want %+v", input, request)
	}
}
