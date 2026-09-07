package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"log"
	"strings"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	emailtemplate "github.com/bayuanugerah/insurance-core-api/internal/templates/email"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
)

const applicationSubmittedSubject = "application.submitted"

type ApplicationService struct {
	products       repositories.ProductRepository
	applications   repositories.ApplicationRepository
	reviewChecks   repositories.ApplicationReviewCheckRepository
	questionnaires repositories.QuestionnaireRepository
	quotes         *ProductService
	mailer         ports.Mailer
	messageBus     ports.MessageBus
	emailRenderer  *emailtemplate.Renderer
}

func NewApplicationService(products repositories.ProductRepository, applications repositories.ApplicationRepository, reviewChecks repositories.ApplicationReviewCheckRepository, quotes *ProductService, mailer ports.Mailer, messageBus ports.MessageBus, questionnaires ...repositories.QuestionnaireRepository) *ApplicationService {
	renderer, err := emailtemplate.NewRenderer()
	if err != nil {
		renderer = nil
	}
	var qRepo repositories.QuestionnaireRepository
	if len(questionnaires) > 0 {
		qRepo = questionnaires[0]
	}
	return &ApplicationService{products: products, applications: applications, reviewChecks: reviewChecks, questionnaires: qRepo, quotes: quotes, mailer: mailer, messageBus: messageBus, emailRenderer: renderer}
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

	var questions []models.Question
	if service.questionnaires != nil && len(input.Answers) > 0 {
		_, qList, err := service.questionnaires.FindByProductIDOrCategory(ctx, product.ID, string(product.Category))
		if err == nil && len(qList) > 0 {
			questions = qList
		}
	}

	totalMultiplier := evaluateQuestionnairePricingMultiplier(questions, input.Answers)

	finalPremium := quote.EstimatedPremium
	if totalMultiplier > 0 && totalMultiplier != 1.0 {
		finalPremium = roundUpToNearestThousand(float64(quote.EstimatedPremium) * totalMultiplier)
	}

	id, err := applicationID()
	if err != nil {
		return models.Application{}, err
	}

	var reviewChecks []models.ApplicationReviewCheck
	if len(questions) > 0 && len(input.Answers) > 0 {
		if err := validations.ValidateAnswers(questions, input.Answers); err != nil {
			return models.Application{}, err
		}
		questionnaireService := NewQuestionnaireService(service.questionnaires, service.products)
		reviewChecks = questionnaireService.EvaluateUnderwritingReviewChecks(id, questions, input.Answers)

		answers := make([]models.ApplicationAnswer, 0, len(input.Answers))
		for _, a := range input.Answers {
			ansID, err := applicationID()
			if err != nil {
				return models.Application{}, err
			}
			answers = append(answers, models.ApplicationAnswer{
				ID:            ansID,
				ApplicationID: id,
				QuestionID:    a.QuestionID,
				Code:          a.Code,
				AnswerValue:   a.Value,
				CreatedAt:     time.Now().UTC(),
			})
		}
		if err := service.questionnaires.SaveAnswers(ctx, answers); err != nil {
			return models.Application{}, err
		}
	}
	if len(reviewChecks) == 0 {
		reviewChecks = defaultApplicationReviewChecks(id)
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
		Premium:          finalPremium,
		Status:           models.ApplicationStatusSubmitted,
		ReviewChecks:     reviewChecks,
	}

	if err := service.applications.Create(ctx, &application); err != nil {
		return models.Application{}, err
	}
	service.publishApplicationSubmitted(ctx, application)
	if service.mailer != nil {
		message, err := service.applicationSubmittedEmail(application, product.Name)
		if err != nil {
			return models.Application{}, err
		}
		if err := service.mailer.Send(ctx, message); err != nil {
			return models.Application{}, err
		}
	}

	return application, nil
}

func (service *ApplicationService) applicationSubmittedEmail(application models.Application, productName string) (ports.EmailMessage, error) {
	textBody := ""
	htmlBody := ""
	if service.emailRenderer != nil {
		var err error
		textBody, htmlBody, err = service.emailRenderer.RenderApplicationSubmitted(emailtemplate.ApplicationSubmittedData{
			FullName:      application.FullName,
			ProductName:   productName,
			ApplicationID: application.ID,
		})
		if err != nil {
			return ports.EmailMessage{}, err
		}
	}

	if textBody == "" {
		textBody = "Halo " + sanitizeEmailText(application.FullName) + ",\n\n" +
			"Pengajuan polis " + sanitizeEmailText(productName) + " berhasil diterima dengan nomor aplikasi " + sanitizeEmailText(application.ID) + ". " +
			"Tim kami akan meninjau data Anda dan menghubungi Anda untuk proses berikutnya.\n\n" +
			"Terima kasih."
	}
	if htmlBody == "" {
		fullName := html.EscapeString(sanitizeEmailText(application.FullName))
		escapedProductName := html.EscapeString(sanitizeEmailText(productName))
		escapedApplicationID := html.EscapeString(sanitizeEmailText(application.ID))
		htmlBody = "<p>Halo " + fullName + ",</p>" +
			"<p>Pengajuan polis <strong>" + escapedProductName + "</strong> berhasil diterima dengan nomor aplikasi <strong>" + escapedApplicationID + "</strong>.</p>" +
			"<p>Tim kami akan meninjau data Anda dan menghubungi Anda untuk proses berikutnya.</p>" +
			"<p>Terima kasih.</p>"
	}

	return ports.EmailMessage{
		To:       []string{application.Email},
		Subject:  "Pengajuan polis berhasil diterima",
		TextBody: textBody,
		HTMLBody: htmlBody,
	}, nil
}

type applicationSubmittedEvent struct {
	ApplicationID string `json:"application_id"`
	ProductID     string `json:"product_id"`
	Email         string `json:"email"`
	Premium       int64  `json:"premium"`
	Status        string `json:"status"`
}

func sanitizeEmailText(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}

func (service *ApplicationService) Get(ctx context.Context, id string) (models.Application, error) {
	if service.applications == nil {
		return models.Application{}, errors.New(constants.ErrApplicationServiceUnavailable)
	}

	return service.applications.FindByID(ctx, id)
}

func (service *ApplicationService) List(ctx context.Context, query dtos.ApplicationListQuery) ([]models.Application, int64, error) {
	if service.applications == nil {
		return nil, 0, errors.New(constants.ErrApplicationServiceUnavailable)
	}

	applications, total, err := service.applications.List(ctx, repositories.ApplicationListFilter{
		Status:    query.Status,
		ProductID: query.ProductID,
		Page:      query.Page,
		Limit:     query.Limit,
	})
	if err != nil {
		return nil, 0, err
	}

	return applications, total, nil
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
	case models.ApplicationStatusDraft:
		return next == models.ApplicationStatusSubmitted
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

func (service *ApplicationService) publishApplicationSubmitted(ctx context.Context, application models.Application) {
	if service.messageBus == nil {
		return
	}
	if err := service.messageBus.PublishJSON(ctx, applicationSubmittedSubject, applicationSubmittedEvent{
		ApplicationID: application.ID,
		ProductID:     application.ProductID,
		Email:         application.Email,
		Premium:       application.Premium,
		Status:        string(application.Status),
	}); err != nil {
		log.Printf("[ApplicationService] warning: failed to publish %s event: %v", applicationSubmittedSubject, err)
	}
}

func evaluateQuestionnairePricingMultiplier(questions []models.Question, answers []dtos.ApplicationAnswerInput) float64 {
	if len(answers) == 0 || len(questions) == 0 {
		return 1.0
	}

	answerMap := make(map[string]any, len(answers))
	for _, a := range inputToMap(answers) {
		answerMap[a.key] = a.val
	}

	totalMultiplier := 1.0

	for _, q := range questions {
		val, exists := answerMap[q.ID]
		if !exists {
			val, exists = answerMap[q.Code]
		}
		if !exists || val == nil {
			continue
		}

		valStr := strings.TrimSpace(fmt.Sprintf("%v", val))
		for _, opt := range q.Options {
			if strings.EqualFold(opt.Value, valStr) {
				if opt.Multiplier > 0 {
					totalMultiplier *= opt.Multiplier
				}
				break
			}
		}
	}

	return totalMultiplier
}

type answerEntry struct {
	key string
	val any
}

func inputToMap(answers []dtos.ApplicationAnswerInput) []answerEntry {
	entries := make([]answerEntry, 0, len(answers)*2)
	for _, a := range answers {
		if k := strings.TrimSpace(a.QuestionID); k != "" {
			entries = append(entries, answerEntry{key: k, val: a.Value})
		}
		if k := strings.TrimSpace(a.Code); k != "" {
			entries = append(entries, answerEntry{key: k, val: a.Value})
		}
	}
	return entries
}
