package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type fakeApplicationRepository struct {
	application     models.Application
	err             error
	created         *models.Application
	updatedID       string
	updatedStatus   models.ApplicationStatus
	updatedReviewer string
	updatedReason   string
	updatedAt       time.Time
}

type fakeReviewCheckRepository struct {
	checks        []models.ApplicationReviewCheck
	err           error
	updatedType   models.ApplicationReviewCheckType
	updatedStatus models.ApplicationReviewCheckStatus
	updatedBy     string
	updatedNotes  string
}

func (repository *fakeApplicationRepository) Create(ctx context.Context, application *models.Application) error {
	repository.created = application
	return repository.err
}

func (repository *fakeApplicationRepository) FindByID(ctx context.Context, id string) (models.Application, error) {
	return repository.application, repository.err
}

func (repository *fakeApplicationRepository) UpdateStatus(ctx context.Context, id string, status models.ApplicationStatus, reviewedBy, rejectionReason string, reviewedAt time.Time) error {
	repository.updatedID = id
	repository.updatedStatus = status
	repository.updatedReviewer = reviewedBy
	repository.updatedReason = rejectionReason
	repository.updatedAt = reviewedAt
	return repository.err
}

func (repository *fakeApplicationRepository) List(ctx context.Context, filter repositories.ApplicationListFilter) ([]models.Application, int64, error) {
	return []models.Application{repository.application}, 1, repository.err
}

func (repository *fakeReviewCheckRepository) FindByApplicationID(ctx context.Context, applicationID string) ([]models.ApplicationReviewCheck, error) {
	return repository.checks, repository.err
}

func (repository *fakeReviewCheckRepository) UpdateStatus(ctx context.Context, applicationID string, checkType models.ApplicationReviewCheckType, status models.ApplicationReviewCheckStatus, reviewedBy, notes string, reviewedAt time.Time) error {
	repository.updatedType = checkType
	repository.updatedStatus = status
	repository.updatedBy = reviewedBy
	repository.updatedNotes = notes
	return repository.err
}

func TestApplicationServiceCreate(t *testing.T) {
	products := &fakeProductRepository{product: productFixture()}
	applications := &fakeApplicationRepository{}
	service := NewApplicationService(products, applications, &fakeReviewCheckRepository{}, NewProductService(products))

	application, err := service.Create(context.Background(), "secure-life-plus", applicationRequestFixture())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if application.ID == "" || len(application.ID) != 32 {
		t.Fatalf("ID = %q, want 32 hex chars", application.ID)
	}
	if application.ProductID != "product-1" || application.Status != models.ApplicationStatusSubmitted {
		t.Fatalf("Create() = %+v, want product-1 submitted", application)
	}
	if application.Premium != 139000 {
		t.Fatalf("Premium = %d, want 139000", application.Premium)
	}
	if applications.created == nil || applications.created.ID != application.ID {
		t.Fatalf("created application = %+v, want persisted application", applications.created)
	}
	if len(applications.created.ReviewChecks) != 4 {
		t.Fatalf("created review checks = %+v, want 4 default checks", applications.created.ReviewChecks)
	}
}

func TestApplicationServiceCreateValidatesDependenciesAndProduct(t *testing.T) {
	_, err := NewApplicationService(nil, nil, nil, nil).Create(context.Background(), "secure-life-plus", applicationRequestFixture())
	if err == nil || err.Error() != constants.ErrApplicationServiceUnavailable {
		t.Fatalf("Create() error = %v, want service unavailable", err)
	}

	products := &fakeProductRepository{product: models.Product{}}
	applications := &fakeApplicationRepository{}
	_, err = NewApplicationService(products, applications, &fakeReviewCheckRepository{}, NewProductService(products)).Create(context.Background(), "missing", applicationRequestFixture())
	if !errors.Is(err, repositories.ErrProductNotFound) {
		t.Fatalf("Create() error = %v, want product not found", err)
	}
}

func TestApplicationServiceCreateReturnsQuoteError(t *testing.T) {
	products := &fakeProductRepository{product: productFixture()}
	request := applicationRequestFixture()
	request.SumAssured = 1

	_, err := NewApplicationService(products, &fakeApplicationRepository{}, &fakeReviewCheckRepository{}, NewProductService(products)).Create(context.Background(), "secure-life-plus", request)
	if !errors.Is(err, constants.QuoteSumAssuredOutOfRangeError) {
		t.Fatalf("Create() error = %v, want quote range error", err)
	}
}

func TestApplicationServiceCreateReturnsCreateError(t *testing.T) {
	expectedErr := errors.New("insert failed")
	products := &fakeProductRepository{product: productFixture()}
	applications := &fakeApplicationRepository{err: expectedErr}

	_, err := NewApplicationService(products, applications, &fakeReviewCheckRepository{}, NewProductService(products)).Create(context.Background(), "secure-life-plus", applicationRequestFixture())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("Create() error = %v, want %v", err, expectedErr)
	}
}

func TestApplicationServiceGet(t *testing.T) {
	expected := models.Application{ID: "application-1"}
	service := NewApplicationService(nil, &fakeApplicationRepository{application: expected}, nil, nil)

	application, err := service.Get(context.Background(), "application-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if application.ID != expected.ID {
		t.Fatalf("Get() = %+v, want %+v", application, expected)
	}
}

func TestApplicationServiceGetRequiresRepository(t *testing.T) {
	_, err := NewApplicationService(nil, nil, nil, nil).Get(context.Background(), "application-1")
	if err == nil || err.Error() != constants.ErrApplicationServiceUnavailable {
		t.Fatalf("Get() error = %v, want service unavailable", err)
	}
}

func TestApplicationServiceUpdateStatus(t *testing.T) {
	tests := []struct {
		name    string
		current models.ApplicationStatus
		input   dtos.UpdateApplicationStatusRequest
		wantErr error
	}{
		{name: "submit to review", current: models.ApplicationStatusSubmitted, input: statusRequest(models.ApplicationStatusUnderReview, "underwriter", "")},
		{name: "review to approved", current: models.ApplicationStatusUnderReview, input: statusRequest(models.ApplicationStatusApproved, "underwriter", "")},
		{name: "review to rejected", current: models.ApplicationStatusUnderReview, input: statusRequest(models.ApplicationStatusRejected, "underwriter", "documents incomplete")},
		{name: "submitted to approved invalid", current: models.ApplicationStatusSubmitted, input: statusRequest(models.ApplicationStatusApproved, "underwriter", ""), wantErr: constants.ErrApplicationStatusTransitionInvalidError},
		{name: "approved is final", current: models.ApplicationStatusApproved, input: statusRequest(models.ApplicationStatusRejected, "underwriter", "bad"), wantErr: constants.ErrApplicationStatusTransitionInvalidError},
		{name: "rejected needs reason", current: models.ApplicationStatusUnderReview, input: statusRequest(models.ApplicationStatusRejected, "underwriter", ""), wantErr: constants.ErrApplicationRejectionReasonRequiredError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakeApplicationRepository{application: models.Application{ID: "application-1", Status: tt.current}}
			service := NewApplicationService(nil, repository, &fakeReviewCheckRepository{checks: passedReviewChecks("application-1")}, nil)

			err := service.UpdateStatus(context.Background(), "application-1", tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("UpdateStatus() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("UpdateStatus() error = %v", err)
			}
			if repository.updatedID != "application-1" || repository.updatedStatus != tt.input.Status || repository.updatedReviewer != tt.input.ReviewedBy || repository.updatedReason != tt.input.RejectionReason {
				t.Fatalf("updated values = %+v", repository)
			}
			if repository.updatedAt.IsZero() {
				t.Fatal("updatedAt is zero")
			}
		})
	}
}

func TestApplicationServiceListReviewChecks(t *testing.T) {
	checks := passedReviewChecks("application-1")
	service := NewApplicationService(nil, &fakeApplicationRepository{application: models.Application{ID: "application-1"}}, &fakeReviewCheckRepository{checks: checks}, nil)

	got, err := service.ListReviewChecks(context.Background(), "application-1")
	if err != nil {
		t.Fatalf("ListReviewChecks() error = %v", err)
	}
	if len(got) != len(checks) {
		t.Fatalf("ListReviewChecks() length = %d, want %d", len(got), len(checks))
	}
}

func TestApplicationServiceUpdateReviewCheck(t *testing.T) {
	reviewChecks := &fakeReviewCheckRepository{}
	service := NewApplicationService(nil, &fakeApplicationRepository{application: models.Application{ID: "application-1"}}, reviewChecks, nil)

	err := service.UpdateReviewCheck(context.Background(), "application-1", models.ApplicationReviewCheckTypeIdentityVerified, dtos.UpdateApplicationReviewCheckRequest{
		Status:     models.ApplicationReviewCheckStatusPassed,
		ReviewedBy: "underwriter",
		Notes:      "ok",
	})
	if err != nil {
		t.Fatalf("UpdateReviewCheck() error = %v", err)
	}
	if reviewChecks.updatedType != models.ApplicationReviewCheckTypeIdentityVerified || reviewChecks.updatedStatus != models.ApplicationReviewCheckStatusPassed || reviewChecks.updatedBy != "underwriter" || reviewChecks.updatedNotes != "ok" {
		t.Fatalf("updated review check = %+v", reviewChecks)
	}
}

func TestApplicationServiceApprovalRequiresCompletedChecklist(t *testing.T) {
	service := NewApplicationService(nil, &fakeApplicationRepository{application: models.Application{ID: "application-1", Status: models.ApplicationStatusUnderReview}}, &fakeReviewCheckRepository{checks: []models.ApplicationReviewCheck{{Status: models.ApplicationReviewCheckStatusPending}}}, nil)

	err := service.UpdateStatus(context.Background(), "application-1", statusRequest(models.ApplicationStatusApproved, "underwriter", ""))
	if !errors.Is(err, constants.ErrApplicationApprovalChecklistIncompleteError) {
		t.Fatalf("UpdateStatus() error = %v, want checklist incomplete", err)
	}
}

func TestApplicationServiceUpdateStatusReturnsRepositoryErrors(t *testing.T) {
	expectedErr := errors.New("db failed")
	service := NewApplicationService(nil, &fakeApplicationRepository{err: expectedErr}, nil, nil)

	err := service.UpdateStatus(context.Background(), "application-1", statusRequest(models.ApplicationStatusUnderReview, "underwriter", ""))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateStatus() error = %v, want %v", err, expectedErr)
	}
}

func TestApplicationServiceUpdateStatusRequiresRepository(t *testing.T) {
	err := NewApplicationService(nil, nil, nil, nil).UpdateStatus(context.Background(), "application-1", statusRequest(models.ApplicationStatusUnderReview, "underwriter", ""))
	if err == nil || err.Error() != constants.ErrApplicationServiceUnavailable {
		t.Fatalf("UpdateStatus() error = %v, want service unavailable", err)
	}
}

func TestApplicationID(t *testing.T) {
	first, err := applicationID()
	if err != nil {
		t.Fatalf("applicationID() error = %v", err)
	}
	second, err := applicationID()
	if err != nil {
		t.Fatalf("applicationID() error = %v", err)
	}
	if len(first) != 32 || len(second) != 32 || first == second {
		t.Fatalf("applicationID() = %q and %q, want unique 32-char IDs", first, second)
	}
}

func applicationRequestFixture() dtos.CreateApplicationRequest {
	return dtos.CreateApplicationRequest{
		FullName: "Bayu Anugerah",
		Email:    "bayu@example.com",
		Phone:    "+628123456789",
		ProductQuoteRequest: dtos.ProductQuoteRequest{
			Age:              35,
			Gender:           constants.GenderMale,
			SumAssured:       300_000_000,
			PaymentTerm:      10,
			PaymentFrequency: constants.PaymentFrequencyMonthly,
			Smoker:           constants.SmokerNo,
			OccupationClass:  constants.OccupationStandard,
			HealthRisk:       constants.HealthRiskLow,
		},
	}
}

func statusRequest(status models.ApplicationStatus, reviewedBy string, rejectionReason string) dtos.UpdateApplicationStatusRequest {
	return dtos.UpdateApplicationStatusRequest{Status: status, ReviewedBy: reviewedBy, RejectionReason: rejectionReason}
}

func passedReviewChecks(applicationID string) []models.ApplicationReviewCheck {
	checks := defaultApplicationReviewChecks(applicationID)
	for index := range checks {
		checks[index].Status = models.ApplicationReviewCheckStatusPassed
	}
	return checks
}
