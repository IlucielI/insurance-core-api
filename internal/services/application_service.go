package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type ApplicationService struct {
	products     repositories.ProductRepository
	applications repositories.ApplicationRepository
	quotes       *ProductService
}

func NewApplicationService(products repositories.ProductRepository, applications repositories.ApplicationRepository, quotes *ProductService) *ApplicationService {
	return &ApplicationService{products: products, applications: applications, quotes: quotes}
}

func (service *ApplicationService) Create(ctx context.Context, slug string, input dtos.CreateApplicationRequest) (models.Application, error) {
	if service.quotes == nil || service.products == nil || service.applications == nil {
		return models.Application{}, errors.New(constants.ErrApplicationServiceUnavailable)
	}

	product, err := service.products.FindBySlug(ctx, slug)
	if err != nil {
		return models.Application{}, err
	}
	if product.ID == "" {
		return models.Application{}, repositories.ErrProductNotFound
	}

	quote, err := service.quotes.CreateProductQuote(ctx, slug, dtos.ProductQuoteRequestToInput(input.ProductQuoteRequest))
	if err != nil {
		return models.Application{}, err
	}

	id, err := applicationID()
	if err != nil {
		return models.Application{}, err
	}

	application := models.Application{
		ID:                id,
		ProductID:         product.ID,
		FullName:          input.FullName,
		Email:             input.Email,
		Phone:             input.Phone,
		Age:               input.Age,
		Gender:            input.Gender,
		SumAssured:        input.SumAssured,
		PaymentTerm:       input.PaymentTerm,
		PaymentFrequency:  input.PaymentFrequency,
		Smoker:            input.Smoker,
		OccupationClass:   input.OccupationClass,
		HealthRisk:        input.HealthRisk,
		Premium:           quote.EstimatedPremium,
		Status:            models.ApplicationStatusSubmitted,
	}

	return application, service.applications.Create(ctx, &application)
}

func (service *ApplicationService) Get(ctx context.Context, id string) (models.Application, error) {
	if service.applications == nil {
		return models.Application{}, errors.New(constants.ErrApplicationServiceUnavailable)
	}

	return service.applications.FindByID(ctx, id)
}

func validApplicationTransition(current, next models.ApplicationStatus) bool {
	switch current {
	case models.ApplicationStatusSubmitted:
		return next == models.ApplicationStatusUnderReview
	case models.ApplicationStatusUnderReview:
		return next == models.ApplicationStatusApproved || next == models.ApplicationStatusRejected
	default:
		return false
	}
}
func (service *ApplicationService) UpdateStatus(ctx context.Context, id string, input dtos.UpdateApplicationStatusRequest) error {
	if service.applications == nil {
		return errors.New(constants.ErrApplicationServiceUnavailable)
	}

	application, err := service.applications.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !validApplicationTransition(application.Status, input.Status) {
		return constants.ErrApplicationStatusTransitionInvalidError
	}
	if input.Status == models.ApplicationStatusRejected && input.RejectionReason == "" {
		return constants.ErrApplicationRejectionReasonRequiredError
	}
	return service.applications.UpdateStatus(ctx, id, input.Status, input.ReviewedBy, input.RejectionReason, time.Now().UTC())
}

func applicationID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return hex.EncodeToString(value), nil
}
