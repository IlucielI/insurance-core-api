package validations

import (
	"errors"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func ValidateKnowledgeCategory(category string) error {
	trimmed := strings.TrimSpace(category)
	if trimmed == "" {
		return errors.New(constants.ErrKnowledgeDocCategoryInvalid)
	}

	err := validation.Validate(trimmed, validation.In(
		string(models.KnowledgeCategoryUnderwriting),
		string(models.KnowledgeCategoryProduct),
		string(models.KnowledgeCategoryClaimFAQ),
		string(models.KnowledgeCategoryCompliance),
		string(models.KnowledgeCategoryCompany),
	))
	if err != nil {
		return errors.New(constants.ErrKnowledgeDocCategoryInvalid)
	}
	return nil
}

func ValidateIndexingStatus(status string) error {
	trimmed := strings.TrimSpace(status)
	if trimmed == "" {
		return errors.New(constants.ErrKnowledgeDocStatusInvalid)
	}

	err := validation.Validate(trimmed, validation.In(
		string(models.IndexingStatusIndexed),
		string(models.IndexingStatusSyncing),
		string(models.IndexingStatusDraft),
	))
	if err != nil {
		return errors.New(constants.ErrKnowledgeDocStatusInvalid)
	}
	return nil
}

func ValidateCreateKnowledgeDocRequest(req *dtos.CreateKnowledgeDocumentRequest) error {
	if req == nil {
		return errors.New(constants.ErrKnowledgeDocRequestBodyInvalid)
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return errors.New(constants.ErrKnowledgeDocTitleRequired)
	}
	if len(req.Title) < 3 || len(req.Title) > 255 {
		return errors.New(constants.ErrKnowledgeDocTitleInvalid)
	}

	if err := ValidateKnowledgeCategory(req.Category); err != nil {
		return err
	}

	req.Summary = strings.TrimSpace(req.Summary)
	if req.Summary == "" {
		return errors.New(constants.ErrKnowledgeDocSummaryRequired)
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		return errors.New(constants.ErrKnowledgeDocContentRequired)
	}

	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	if req.Slug != "" {
		if !slugRegex.MatchString(req.Slug) || len(req.Slug) > 255 {
			return errors.New(constants.ErrKnowledgeDocSlugInvalid)
		}
	}

	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if req.Status != "" {
		if err := ValidateIndexingStatus(req.Status); err != nil {
			return err
		}
	} else {
		req.Status = string(models.IndexingStatusIndexed)
	}

	if req.Tags == nil {
		req.Tags = []string{}
	} else {
		var cleanedTags []string
		for _, tag := range req.Tags {
			t := strings.ToLower(strings.TrimSpace(tag))
			if t != "" {
				cleanedTags = append(cleanedTags, t)
			}
		}
		req.Tags = cleanedTags
	}

	return nil
}

func ValidateUpdateKnowledgeDocRequest(req *dtos.UpdateKnowledgeDocumentRequest) error {
	if req == nil {
		return errors.New(constants.ErrKnowledgeDocRequestBodyInvalid)
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return errors.New(constants.ErrKnowledgeDocTitleRequired)
		}
		if len(t) < 3 || len(t) > 255 {
			return errors.New(constants.ErrKnowledgeDocTitleInvalid)
		}
		*req.Title = t
	}

	if req.Category != nil {
		c := strings.TrimSpace(*req.Category)
		if err := ValidateKnowledgeCategory(c); err != nil {
			return err
		}
		*req.Category = c
	}

	if req.Summary != nil {
		s := strings.TrimSpace(*req.Summary)
		if s == "" {
			return errors.New(constants.ErrKnowledgeDocSummaryRequired)
		}
		*req.Summary = s
	}

	if req.Content != nil {
		c := strings.TrimSpace(*req.Content)
		if c == "" {
			return errors.New(constants.ErrKnowledgeDocContentRequired)
		}
		*req.Content = c
	}

	if req.Slug != nil {
		s := strings.ToLower(strings.TrimSpace(*req.Slug))
		if s == "" || !slugRegex.MatchString(s) || len(s) > 255 {
			return errors.New(constants.ErrKnowledgeDocSlugInvalid)
		}
		*req.Slug = s
	}

	if req.Status != nil {
		st := strings.ToLower(strings.TrimSpace(*req.Status))
		if err := ValidateIndexingStatus(st); err != nil {
			return err
		}
		*req.Status = st
	}

	if req.Tags != nil {
		var cleanedTags []string
		for _, tag := range *req.Tags {
			t := strings.ToLower(strings.TrimSpace(tag))
			if t != "" {
				cleanedTags = append(cleanedTags, t)
			}
		}
		*req.Tags = cleanedTags
	}

	return nil
}

func ValidateKnowledgeListQuery(category, status string) (string, string, error) {
	cat := strings.TrimSpace(category)
	if cat != "" && cat != "all" {
		if err := ValidateKnowledgeCategory(cat); err != nil {
			return "", "", err
		}
	} else if cat == "all" {
		cat = ""
	}

	st := strings.TrimSpace(status)
	if st != "" && st != "all" {
		if err := ValidateIndexingStatus(st); err != nil {
			return "", "", err
		}
	} else if st == "all" {
		st = ""
	}

	return cat, st, nil
}
