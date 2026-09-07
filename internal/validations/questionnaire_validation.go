package validations

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

// ValidateAnswers validates input answers against active questions in the questionnaire.
func ValidateAnswers(questions []models.Question, answers []dtos.ApplicationAnswerInput) error {
	answersMap := make(map[string]any, len(answers))
	for _, ans := range answers {
		key := strings.TrimSpace(ans.QuestionID)
		if key == "" {
			key = strings.TrimSpace(ans.Code)
		}
		if key != "" {
			answersMap[key] = ans.Value
			answersMap[strings.TrimSpace(ans.Code)] = ans.Value
		}
	}

	for _, q := range questions {
		if !q.IsActive {
			continue
		}

		// Check parent condition
		if q.ParentQuestionID != nil && *q.ParentQuestionID != "" {
			parentVal, exists := answersMap[*q.ParentQuestionID]
			if !exists || !matchesParentCondition(parentVal, q.ShowIfParentValue) {
				// Dependent question is not active, skip validation
				continue
			}
		}

		val, hasAnswer := answersMap[q.ID]
		if !hasAnswer {
			val, hasAnswer = answersMap[q.Code]
		}

		// Check required rule
		if isRequired(q.ValidationRules) {
			if !hasAnswer || isValueEmpty(val) {
				errMsg, ok := q.ValidationRules["error_message"].(string)
				if !ok || errMsg == "" {
					errMsg = fmt.Sprintf("pertanyaan %s wajib diisi", q.Label)
				}
				return errors.New(errMsg)
			}
		}

		if hasAnswer && !isValueEmpty(val) {
			if err := validateValueRules(q, val); err != nil {
				return err
			}
		}
	}

	return nil
}

func isRequired(rules map[string]any) bool {
	if rules == nil {
		return false
	}
	req, ok := rules["required"].(bool)
	return ok && req
}

func isValueEmpty(val any) bool {
	if val == nil {
		return true
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case bool:
		return false
	case float64:
		return false
	case int, int64:
		return false
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}

func matchesParentCondition(actual any, expected any) bool {
	if actual == nil || expected == nil {
		return false
	}
	actualStr := strings.TrimSpace(fmt.Sprintf("%v", actual))
	expectedStr := strings.TrimSpace(fmt.Sprintf("%v", expected))
	return strings.EqualFold(actualStr, expectedStr)
}

func validateValueRules(q models.Question, val any) error {
	rules := q.ValidationRules
	if rules == nil {
		return nil
	}

	strVal := fmt.Sprintf("%v", val)

	// Pattern validation
	if pattern, ok := rules["pattern"].(string); ok && pattern != "" {
		matched, err := regexp.MatchString(pattern, strVal)
		if err != nil || !matched {
			if msg, hasMsg := rules["error_message"].(string); hasMsg && msg != "" {
				return errors.New(msg)
			}
			return fmt.Errorf("%s tidak memenuhi format yang valid", q.Label)
		}
	}

	// String length validation
	if minLen, ok := rules["min_length"].(float64); ok {
		if len(strVal) < int(minLen) {
			return fmt.Errorf("%s minimal %d karakter", q.Label, int(minLen))
		}
	}
	if maxLen, ok := rules["max_length"].(float64); ok {
		if len(strVal) > int(maxLen) {
			return fmt.Errorf("%s maksimal %d karakter", q.Label, int(maxLen))
		}
	}

	return nil
}
