package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
)

type QuestionnaireService struct {
	questionnaires repositories.QuestionnaireRepository
	products       repositories.ProductRepository
}

func NewQuestionnaireService(
	questionnaires repositories.QuestionnaireRepository,
	products repositories.ProductRepository,
) *QuestionnaireService {
	return &QuestionnaireService{
		questionnaires: questionnaires,
		products:       products,
	}
}

func (s *QuestionnaireService) GetByProductSlug(ctx context.Context, slug string) (*dtos.QuestionnaireResponse, error) {
	var productID string
	var category string
	var productSlug string

	trimmedSlug := strings.TrimSpace(slug)
	if trimmedSlug != "" && s.products != nil {
		product, err := s.products.FindBySlug(ctx, trimmedSlug)
		if err == nil && product.ID != "" {
			productID = product.ID
			category = string(product.Category)
			productSlug = product.Slug
		}
	}

	questionnaire, questions, err := s.questionnaires.FindByProductIDOrCategory(ctx, productID, category)
	if err != nil {
		return nil, err
	}

	return s.buildResponse(questionnaire, questions, productSlug), nil
}

func (s *QuestionnaireService) GetDefault(ctx context.Context) (*dtos.QuestionnaireResponse, error) {
	questionnaire, questions, err := s.questionnaires.FindByProductIDOrCategory(ctx, "", "all")
	if err != nil {
		return nil, err
	}

	return s.buildResponse(questionnaire, questions, ""), nil
}

func (s *QuestionnaireService) ValidateAnswers(questions []models.Question, answers []dtos.ApplicationAnswerInput) error {
	return validations.ValidateAnswers(questions, answers)
}

func (s *QuestionnaireService) EvaluateUnderwritingReviewChecks(
	applicationID string,
	questions []models.Question,
	answers []dtos.ApplicationAnswerInput,
) []models.ApplicationReviewCheck {
	// Base default checks
	checkTypes := []models.ApplicationReviewCheckType{
		models.ApplicationReviewCheckTypeIdentityVerified,
		models.ApplicationReviewCheckTypeIncomeVerified,
		models.ApplicationReviewCheckTypeMedicalRequired,
		models.ApplicationReviewCheckTypeDocumentsComplete,
	}

	checkMap := make(map[models.ApplicationReviewCheckType]*models.ApplicationReviewCheck, len(checkTypes))
	checkMap[models.ApplicationReviewCheckTypeIdentityVerified] = &models.ApplicationReviewCheck{
		ID:            fmt.Sprintf("%s-%s", applicationID, models.ApplicationReviewCheckTypeIdentityVerified),
		ApplicationID: applicationID,
		CheckType:     models.ApplicationReviewCheckTypeIdentityVerified,
		Status:        models.ApplicationReviewCheckStatusPassed,
		Notes:         "Identitas Dukcapil terverifikasi valid",
	}
	checkMap[models.ApplicationReviewCheckTypeIncomeVerified] = &models.ApplicationReviewCheck{
		ID:            fmt.Sprintf("%s-%s", applicationID, models.ApplicationReviewCheckTypeIncomeVerified),
		ApplicationID: applicationID,
		CheckType:     models.ApplicationReviewCheckTypeIncomeVerified,
		Status:        models.ApplicationReviewCheckStatusPassed,
		Notes:         "Kapasitas finansial sehat (DSR dalam batas aman OJK)",
	}
	checkMap[models.ApplicationReviewCheckTypeMedicalRequired] = &models.ApplicationReviewCheck{
		ID:            fmt.Sprintf("%s-%s", applicationID, models.ApplicationReviewCheckTypeMedicalRequired),
		ApplicationID: applicationID,
		CheckType:     models.ApplicationReviewCheckTypeMedicalRequired,
		Status:        models.ApplicationReviewCheckStatusNotNeeded,
		Notes:         "Bebas riwayat medis berisiko (Medical Exam Waived)",
	}
	checkMap[models.ApplicationReviewCheckTypeDocumentsComplete] = &models.ApplicationReviewCheck{
		ID:            fmt.Sprintf("%s-%s", applicationID, models.ApplicationReviewCheckTypeDocumentsComplete),
		ApplicationID: applicationID,
		CheckType:     models.ApplicationReviewCheckTypeDocumentsComplete,
		Status:        models.ApplicationReviewCheckStatusPassed,
		Notes:         "Dokumen legalitas dan ahli waris lengkap",
	}

	// Build lookup for answers
	answerMap := make(map[string]any, len(answers))
	for _, a := range answers {
		answerMap[a.QuestionID] = a.Value
		answerMap[a.Code] = a.Value
	}

	for _, q := range questions {
		val, exists := answerMap[q.ID]
		if !exists {
			val, exists = answerMap[q.Code]
		}
		if !exists || val == nil {
			continue
		}

		// 1. Evaluate underwriting rules from database question
		if q.UnderwritingRules != nil {
			flagIf, ok := q.UnderwritingRules["flag_if"].(string)
			if ok && flagIf != "" {
				valStr := strings.TrimSpace(fmt.Sprintf("%v", val))
				if strings.EqualFold(valStr, flagIf) {
					targetPillar := models.ApplicationReviewCheckType(q.PillarType)
					if check, found := checkMap[targetPillar]; found {
						status := models.ApplicationReviewCheckStatusPending
						if ps, ok := q.UnderwritingRules["pillar_status"].(string); ok && ps != "" {
							status = models.ApplicationReviewCheckStatus(ps)
						}
						check.Status = status
						rfiType, ok := q.UnderwritingRules["rfi_type"].(string)
						if ok && rfiType != "" {
							check.Notes = fmt.Sprintf("Perlu klarifikasi: %s (Dokumen yang diperlukan: %s)", q.Label, rfiType)
						} else {
							check.Notes = fmt.Sprintf("Catatan underwriting khusus: %s", q.Label)
						}
					}
				}
			}
		}

		// 2. Evaluate risk actions on selected option from database
		valStr := strings.TrimSpace(fmt.Sprintf("%v", val))
		for _, opt := range q.Options {
			if strings.EqualFold(opt.Value, valStr) {
				if opt.RiskImpact != "" || opt.RfiRequired {
					targetPillar := models.ApplicationReviewCheckType(q.PillarType)
					if check, found := checkMap[targetPillar]; found {
						check.Status = models.ApplicationReviewCheckStatusPending
						if opt.RfiRequired && check.Notes == "" {
							check.Notes = fmt.Sprintf("Perlu klarifikasi: %s (%s)", q.Label, opt.Label)
						}
					}
				}
			}
		}
	}

	results := make([]models.ApplicationReviewCheck, 0, len(checkTypes))
	for _, ct := range checkTypes {
		if c, ok := checkMap[ct]; ok {
			results = append(results, *c)
		}
	}

	return results
}

func (s *QuestionnaireService) buildResponse(
	q *models.Questionnaire,
	questions []models.Question,
	productSlug string,
) *dtos.QuestionnaireResponse {
	stepMetadata := map[int]struct {
		pillarType  string
		title       string
		description string
	}{
		1: {
			pillarType:  "identity_verified",
			title:       "Identitas Dukcapil",
			description: "Verifikasi NIK e-KTP, nama lengkap, dan data biometrik kependudukan resmi.",
		},
		2: {
			pillarType:  "income_verified",
			title:       "Finansial & Rasio DSR",
			description: "Analisis kemampuan pembayaran premi berkesinambungan tanpa risiko gagal bayar.",
		},
		3: {
			pillarType:  "medical_required",
			title:       "Skrining Medis & Gaya Hidup",
			description: "Kuesioner penyakit kritis, riwayat rawat inap, dan indeks massa tubuh (BMI).",
		},
		4: {
			pillarType:  "documents_complete",
			title:       "Legalitas & Beneficiary",
			description: "Penunjukan ahli waris sah dan persetujuan klausul e-Policy digital.",
		},
	}

	stepMap := make(map[int][]models.Question)
	for _, item := range questions {
		stepMap[item.StepNumber] = append(stepMap[item.StepNumber], item)
	}

	steps := make([]dtos.QuestionnaireStepGroup, 0, 4)
	for i := 1; i <= 4; i++ {
		meta := stepMetadata[i]
		qList := stepMap[i]
		if qList == nil {
			qList = []models.Question{}
		}
		steps = append(steps, dtos.QuestionnaireStepGroup{
			StepNumber:  i,
			PillarType:  meta.pillarType,
			Title:       meta.title,
			Description: meta.description,
			Questions:   qList,
		})
	}

	return &dtos.QuestionnaireResponse{
		ID:          q.ID,
		ProductID:   q.ProductID,
		ProductSlug: productSlug,
		Category:    q.Category,
		Title:       q.Title,
		Description: q.Description,
		Version:     q.Version,
		Steps:       steps,
		Questions:   questions,
	}
}
