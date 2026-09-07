package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

func TestPostgresPricingRuleRepositoryWithCache(t *testing.T) {
	cache := newFakeCache()

	rules := []models.ProductPricingRule{
		{
			ID:        "rule-1",
			ProductID: "prod-1",
			RuleCode:  "base_rate",
			RuleName:  "Base Rate",
			RuleType:  "base_rate",
			Factors:   map[string]any{"rate": 0.0035},
			IsActive:  true,
		},
		{
			ID:        "rule-2",
			ProductID: "prod-1",
			RuleCode:  "age_bracket",
			RuleName:  "Age Bracket",
			RuleType:  "bracket",
			Factors: map[string]any{
				"brackets": []any{
					map[string]any{"min_age": 18, "max_age": 30, "factor": 1.0},
					map[string]any{"min_age": 31, "max_age": 50, "factor": 1.2},
				},
			},
			IsActive: true,
		},
	}

	productIDCacheKey := "tenant:global:pricing_rules:product_id:prod-1"
	if err := cache.SetJSON(context.Background(), productIDCacheKey, rules, time.Hour); err != nil {
		t.Fatalf("cache.SetJSON error = %v", err)
	}

	slugCacheKey := "tenant:global:pricing_rules:product_slug:secure-life"
	if err := cache.SetJSON(context.Background(), slugCacheKey, rules, time.Hour); err != nil {
		t.Fatalf("cache.SetJSON error = %v", err)
	}

	repo := NewPostgresPricingRuleRepository(nil).WithCache(cache)

	// 1. FindByProductID hits cache
	foundRules, err := repo.FindByProductID(context.Background(), "prod-1")
	if err != nil {
		t.Fatalf("FindByProductID() error = %v", err)
	}
	if len(foundRules) != 2 {
		t.Fatalf("len(foundRules) = %d, want 2", len(foundRules))
	}
	if foundRules[0].RuleCode != "base_rate" {
		t.Fatalf("foundRules[0].RuleCode = %q, want base_rate", foundRules[0].RuleCode)
	}

	// 2. FindByProductSlug hits cache
	foundBySlug, err := repo.FindByProductSlug(context.Background(), "secure-life")
	if err != nil {
		t.Fatalf("FindByProductSlug() error = %v", err)
	}
	if len(foundBySlug) != 2 {
		t.Fatalf("len(foundBySlug) = %d, want 2", len(foundBySlug))
	}
	if foundBySlug[1].RuleCode != "age_bracket" {
		t.Fatalf("foundBySlug[1].RuleCode = %q, want age_bracket", foundBySlug[1].RuleCode)
	}
}
