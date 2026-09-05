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
	reviewChecks repositories.ApplicationReviewCheckRepository
	quotes       *ProductService
}

func NewApplicationService(products repositories.ProductRepository, applications repositories.ApplicationRepository, reviewChecks repositories.ApplicationReviewCheckRepository, quotes *ProductService) *ApplicationService {
	return &ApplicationService{products: products, applications: applications, reviewChecks: reviewChecks, quotes: quotes}
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
		ID:               id,
		ProductID:        product.ID,
		FullName:         input.FullName,
		Email:            input.Email,
		Phone:            input.Phone,
		Age:              input.Age,
		Gender:           input.Gender,
		SumAssured:       input.SumAssured,
		PaymentTerm:      input.PaymentTerm,
		PaymentFrequency: input.PaymentFrequency,
		Smoker:           input.Smoker,
		OccupationClass:  input.OccupationClass,
		HealthRisk:       input.HealthRisk,
		Premium:          quote.EstimatedPremium,
		Status:           models.ApplicationStatusSubmitted,
		ReviewChecks:     defaultApplicationReviewChecks(id),
	}

	return application, service.applications.Create(ctx, &application)
}

func (service *ApplicationService) Get(ctx context.Context, id string) (models.Application, error) {
	if service.applications == nil {
		return models.Application{}, errors.New(constants.ErrApplicationServiceUnavailable)
	}

	return service.applications.FindByID(ctx, id)
}

func (service *ApplicationService) ListReviewChecks(ctx context.Context, applicationID string) ([]models.ApplicationReviewCheck, error) {
	if service.applications == nil || service.reviewChecks == nil {
		return nil, errors.New(constants.ErrApplicationServiceUnavailable)
	}

	if _, err := service.applications.FindByID(ctx, applicationID); err != nil {
		return nil, err
	}

	return service.reviewChecks.FindByApplicationID(ctx, applicationID)
}

func (service *ApplicationService) UpdateReviewCheck(ctx context.Context, applicationID string, checkType models.ApplicationReviewCheckType, input dtos.UpdateApplicationReviewCheckRequest) error {
	if service.applications == nil || service.reviewChecks == nil {
		return errors.New(constants.ErrApplicationServiceUnavailable)
	}

	if _, err := service.applications.FindByID(ctx, applicationID); err != nil {
		return err
	}

	return service.reviewChecks.UpdateStatus(ctx, applicationID, checkType, input.Status, input.ReviewedBy, input.Notes, time.Now().UTC())
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
	if input.Status == models.ApplicationStatusApproved {
		complete, err := service.applicationChecklistComplete(ctx, id)
		if err != nil {
			return err
		}
		if !complete {
			return constants.ErrApplicationApprovalChecklistIncompleteError
		}
	}
	if input.Status == models.ApplicationStatusRejected && input.RejectionReason == "" {
		return constants.ErrApplicationRejectionReasonRequiredError
	}
	return service.applications.UpdateStatus(ctx, id, input.Status, input.ReviewedBy, input.RejectionReason, time.Now().UTC())
}

func (service *ApplicationService) applicationChecklistComplete(ctx context.Context, applicationID string) (bool, error) {
	if service.reviewChecks == nil {
		return false, errors.New(constants.ErrApplicationServiceUnavailable)
	}

	checks, err := service.reviewChecks.FindByApplicationID(ctx, applicationID)
	if err != nil {
		return false, err
	}
	if len(checks) == 0 {
		return false, nil
	}

	for _, check := range checks {
		if check.Status != models.ApplicationReviewCheckStatusPassed && check.Status != models.ApplicationReviewCheckStatusNotNeeded {
			return false, nil
		}
	}

	return true, nil
}

func defaultApplicationReviewChecks(applicationID string) []models.ApplicationReviewCheck {
	checkTypes := []models.ApplicationReviewCheckType{
		models.ApplicationReviewCheckTypeIdentityVerified,
		models.ApplicationReviewCheckTypeIncomeVerified,
		models.ApplicationReviewCheckTypeDocumentsComplete,
		models.ApplicationReviewCheckTypeMedicalRequired,
	}

	checks := make([]models.ApplicationReviewCheck, 0, len(checkTypes))
	for _, checkType := range checkTypes {
		checks = append(checks, models.ApplicationReviewCheck{
			ID:            reviewCheckID(applicationID, checkType),
			ApplicationID: applicationID,
			CheckType:     checkType,
			Status:        models.ApplicationReviewCheckStatusPending,
		})
	}

	return checks
}

func reviewCheckID(applicationID string, checkType models.ApplicationReviewCheckType) string {
	return applicationID + "-" + string(checkType)
}

func applicationID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return hex.EncodeToString(value), nil
}
