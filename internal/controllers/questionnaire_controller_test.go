package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/routes"
)

type controllerQuestionnaireRepository struct {
	questionnaire *models.Questionnaire
	questions     []models.Question
	savedAnswers  []models.ApplicationAnswer
	err           error
}

func (r *controllerQuestionnaireRepository) FindByProductIDOrCategory(ctx context.Context, productID string, category string) (*models.Questionnaire, []models.Question, error) {
	if r.err != nil {
		return nil, nil, r.err
	}
	if r.questionnaire == nil {
		return nil, nil, repositories.ErrQuestionnaireNotFound
	}
	return r.questionnaire, r.questions, nil
}

func (r *controllerQuestionnaireRepository) FindQuestionsByQuestionnaireID(ctx context.Context, questionnaireID string) ([]models.Question, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.questions, nil
}

func (r *controllerQuestionnaireRepository) SaveAnswers(ctx context.Context, answers []models.ApplicationAnswer) error {
	if r.err != nil {
		return r.err
	}
	r.savedAnswers = append(r.savedAnswers, answers...)
	return nil
}

func (r *controllerQuestionnaireRepository) FindAnswersByApplicationID(ctx context.Context, applicationID string) ([]models.ApplicationAnswer, error) {
	return r.savedAnswers, nil
}

func TestQuestionnaireController_Endpoints(t *testing.T) {
	product := models.Product{
		ID:              "prod_secure_life",
		Name:            "Secure Life Plus",
		Slug:            "secure-life-plus",
		Category:        models.ProductCategoryLife,
		MinSumAssured:   100_000_000,
		MaxSumAssured:   1_000_000_000,
		MinPaymentTerm:  5,
		MaxPaymentTerm:  20,
		StartingPremium: 100_000,
		PricingRules: models.PricingRules{
			BaseRate: 0.0035,
			AgeFactors: []models.AgeFactor{
				{MinAge: 18, MaxAge: 60, Factor: 1.0},
			},
			GenderFactors:     map[string]float64{"male": 1.0, "female": 1.0},
			SmokerFactors:     map[string]float64{"yes": 1.0, "no": 1.0},
			OccupationFactors: map[string]float64{"low": 1.0, "standard": 1.0, "high": 1.0},
			HealthFactors:     map[string]float64{"low": 1.0, "medium": 1.0, "high": 1.0, "standard": 1.0},
			FrequencyLoading:  map[string]float64{"monthly": 1.0, "annual": 1.0},
		},
	}

	qRepo := &controllerQuestionnaireRepository{
		questionnaire: &models.Questionnaire{
			ID:       "quest_default",
			Title:    "Standar Underwriting",
			Version:  1,
			Category: "all",
			IsActive: true,
		},
		questions: []models.Question{
			{
				ID:         "q_nik",
				StepNumber: 1,
				PillarType: "identity_verified",
				Code:       "nik",
				Label:      "NIK",
				InputType:  "text",
				IsActive:   true,
			},
			{
				ID:                  "q_smoker",
				StepNumber:          3,
				PillarType:          "medical_required",
				Code:                "is_smoker",
				Label:               "Status Merokok",
				InputType:           "radio",
				AffectsPricingField: "smoker",
				Options: []models.QuestionOption{
					{Value: "no", Label: "Bukan Perokok", Multiplier: 1.0},
					{Value: "yes", Label: "Perokok Aktif", Multiplier: 1.25},
				},
				IsActive: true,
			},
		},
	}

	prodRepo := &controllerProductRepository{product: product}
	appRepo := &controllerApplicationRepository{}
	reviewRepo := &controllerReviewCheckRepository{}

	app := routes.NewRouter(
		config.Config{AppName: "test", Version: "1.0.0", GitHash: "abc"},
		prodRepo,
		appRepo,
		reviewRepo,
		nil,
		nil,
		nil,
		nil,
		qRepo,
	)

	// 1. Test GET /api/v1/products/:slug/questionnaire
	req, err := http.NewRequest(http.MethodGet, "/api/v1/products/secure-life-plus/questionnaire", nil)
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("GET /api/v1/products/:slug/questionnaire failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want 200", resp.StatusCode)
	}

	// 2. Test GET /api/v1/questionnaires
	req, err = http.NewRequest(http.MethodGet, "/api/v1/questionnaires", nil)
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("GET /api/v1/questionnaires failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want 200", resp.StatusCode)
	}

	// 3. Test POST /api/v1/applications (slug resolved from request body, smoker synced from DB question definition)
	payload := dtos.CreateApplicationRequest{
		ProductSlug: "secure-life-plus",
		FullName:    "Budi Santoso",
		Email:       "budi@example.com",
		Phone:       "081234567890",
		ProductQuoteRequest: dtos.ProductQuoteRequest{
			Age:              30,
			Gender:           "male",
			SumAssured:       500_000_000,
			PaymentTerm:      10,
			PaymentFrequency: "monthly",
			Smoker:           "no", // initially no
			OccupationClass:  "standard",
			HealthRisk:       "low",
		},
		Answers: []dtos.ApplicationAnswerInput{
			{QuestionID: "q_nik", Code: "nik", Value: "3201123456780001"},
			{QuestionID: "q_smoker", Code: "is_smoker", Value: "yes"}, // answered yes
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	req, err = http.NewRequest(http.MethodPost, "/api/v1/applications", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("POST /api/v1/applications failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		t.Fatalf("status code = %d, want 201, body = %s", resp.StatusCode, buf.String())
	}
	if appRepo.created == nil {
		t.Fatal("application was not created")
	}
	if appRepo.created.FullName != "Budi Santoso" {
		t.Errorf("created application FullName = %s, want Budi Santoso", appRepo.created.FullName)
	}
	if appRepo.created.Premium <= 0 {
		t.Errorf("created application Premium = %d, want > 0", appRepo.created.Premium)
	}
}
