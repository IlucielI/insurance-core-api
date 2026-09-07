package services_test

import (
	"context"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
)

type fakeQuestionnaireRepository struct {
	questionnaire *models.Questionnaire
	questions     []models.Question
	savedAnswers  []models.ApplicationAnswer
	err           error
}

func (f *fakeQuestionnaireRepository) FindByProductIDOrCategory(ctx context.Context, productID string, category string) (*models.Questionnaire, []models.Question, error) {
	if f.err != nil {
		return nil, nil, f.err
	}
	if f.questionnaire == nil {
		return nil, nil, repositories.ErrQuestionnaireNotFound
	}
	return f.questionnaire, f.questions, nil
}

func (f *fakeQuestionnaireRepository) FindQuestionsByQuestionnaireID(ctx context.Context, questionnaireID string) ([]models.Question, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.questions, nil
}

func (f *fakeQuestionnaireRepository) SaveAnswers(ctx context.Context, answers []models.ApplicationAnswer) error {
	if f.err != nil {
		return f.err
	}
	f.savedAnswers = append(f.savedAnswers, answers...)
	return nil
}

func (f *fakeQuestionnaireRepository) FindAnswersByApplicationID(ctx context.Context, applicationID string) ([]models.ApplicationAnswer, error) {
	return f.savedAnswers, nil
}

func TestQuestionnaireService_GetByProductSlug(t *testing.T) {
	parentID := "q_has_ci"
	qRepo := &fakeQuestionnaireRepository{
		questionnaire: &models.Questionnaire{
			ID:       "quest_1",
			Title:    "Standar Underwriting",
			Version:  1,
			Category: "life",
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
				ID:                "q_has_ci",
				StepNumber:        3,
				PillarType:        "medical_required",
				Code:              "has_ci",
				Label:             "Riwayat Penyakit Kritis",
				InputType:         "radio",
				UnderwritingRules: map[string]any{"flag_if": "yes", "rfi_type": "Resume Medis"},
				IsActive:          true,
			},
			{
				ID:                "q_ci_details",
				StepNumber:        3,
				PillarType:        "medical_required",
				Code:              "ci_details",
				Label:             "Rincian Penyakit Kritis",
				InputType:         "text",
				ParentQuestionID:  &parentID,
				ShowIfParentValue: "yes",
				IsActive:          true,
			},
		},
	}

	svc := services.NewQuestionnaireService(qRepo, nil)
	resp, err := svc.GetByProductSlug(context.Background(), "secure-life-plus")
	if err != nil {
		t.Fatalf("GetByProductSlug() error = %v", err)
	}

	if resp.ID != "quest_1" {
		t.Errorf("resp.ID = %s, want quest_1", resp.ID)
	}
	if len(resp.Steps) != 4 {
		t.Errorf("len(resp.Steps) = %d, want 4", len(resp.Steps))
	}
	if len(resp.Steps[0].Questions) != 1 {
		t.Errorf("len(resp.Steps[0].Questions) = %d, want 1", len(resp.Steps[0].Questions))
	}
	if len(resp.Steps[2].Questions) != 2 {
		t.Errorf("len(resp.Steps[2].Questions) = %d, want 2", len(resp.Steps[2].Questions))
	}
}

func TestQuestionnaireService_ValidateAnswers_And_EvaluateChecks(t *testing.T) {
	parentID := "q_has_ci"
	questions := []models.Question{
		{
			ID:              "q_nik",
			StepNumber:      1,
			PillarType:      "identity_verified",
			Code:            "nik",
			Label:           "NIK",
			ValidationRules: map[string]any{"required": true, "pattern": "^[0-9]{16}$"},
			IsActive:        true,
		},
		{
			ID:                "q_has_ci",
			StepNumber:        3,
			PillarType:        "medical_required",
			Code:              "has_ci",
			Label:             "Riwayat Penyakit Kritis",
			UnderwritingRules: map[string]any{"flag_if": "yes", "rfi_type": "Resume Medis"},
			IsActive:          true,
		},
		{
			ID:                "q_ci_details",
			StepNumber:        3,
			PillarType:        "medical_required",
			Code:              "ci_details",
			Label:             "Rincian Penyakit Kritis",
			ParentQuestionID:  &parentID,
			ShowIfParentValue: "yes",
			ValidationRules:   map[string]any{"required": true, "min_length": float64(5)},
			IsActive:          true,
		},
	}

	svc := services.NewQuestionnaireService(&fakeQuestionnaireRepository{}, nil)

	// Case 1: Invalid NIK (not 16 digits)
	err := svc.ValidateAnswers(questions, []dtos.ApplicationAnswerInput{
		{QuestionID: "q_nik", Code: "nik", Value: "123"},
	})
	if err == nil {
		t.Fatal("expected error for invalid NIK pattern, got nil")
	}

	// Case 2: Valid NIK, has_ci = "no" -> parent is "no", so ci_details is not required
	err = svc.ValidateAnswers(questions, []dtos.ApplicationAnswerInput{
		{QuestionID: "q_nik", Code: "nik", Value: "3201123456780001"},
		{QuestionID: "q_has_ci", Code: "has_ci", Value: "no"},
	})
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	// Case 3: has_ci = "yes" but ci_details missing -> should fail validation
	err = svc.ValidateAnswers(questions, []dtos.ApplicationAnswerInput{
		{QuestionID: "q_nik", Code: "nik", Value: "3201123456780001"},
		{QuestionID: "q_has_ci", Code: "has_ci", Value: "yes"},
	})
	if err == nil {
		t.Fatal("expected error for missing ci_details when parent is yes, got nil")
	}

	// Case 4: has_ci = "yes" with valid ci_details -> evaluate checks should flag medical_required with note
	answers := []dtos.ApplicationAnswerInput{
		{QuestionID: "q_nik", Code: "nik", Value: "3201123456780001"},
		{QuestionID: "q_has_ci", Code: "has_ci", Value: "yes"},
		{QuestionID: "q_ci_details", Code: "ci_details", Value: "Kanker stadium 1 telah operasi"},
	}
	if err := svc.ValidateAnswers(questions, answers); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := svc.EvaluateUnderwritingReviewChecks("app_test_1", questions, answers)
	if len(checks) != 4 {
		t.Fatalf("len(checks) = %d, want 4", len(checks))
	}
	var medCheck *models.ApplicationReviewCheck
	for i := range checks {
		if checks[i].CheckType == models.ApplicationReviewCheckTypeMedicalRequired {
			medCheck = &checks[i]
		}
	}
	if medCheck == nil {
		t.Fatal("medical_required check not found")
	}
	if medCheck.Notes == "" {
		t.Errorf("medCheck.Notes should contain RFI notes, got empty")
	}
	if medCheck.Status != models.ApplicationReviewCheckStatusPending {
		t.Errorf("flagged medCheck.Status = %v, want pending", medCheck.Status)
	}

	// Case 5: has_ci = "no" -> clean, medical_required status should be not_needed (MCU waived)
	cleanAnswers := []dtos.ApplicationAnswerInput{
		{QuestionID: "q_nik", Code: "nik", Value: "3201123456780001"},
		{QuestionID: "q_has_ci", Code: "has_ci", Value: "no"},
	}
	cleanChecks := svc.EvaluateUnderwritingReviewChecks("app_test_2", questions, cleanAnswers)
	var cleanMedCheck *models.ApplicationReviewCheck
	for i := range cleanChecks {
		if cleanChecks[i].CheckType == models.ApplicationReviewCheckTypeMedicalRequired {
			cleanMedCheck = &cleanChecks[i]
		}
	}
	if cleanMedCheck == nil {
		t.Fatal("cleanMedCheck not found")
	}
	if cleanMedCheck.Status != models.ApplicationReviewCheckStatusNotNeeded {
		t.Errorf("cleanMedCheck.Status = %v, want not_needed", cleanMedCheck.Status)
	}
}
