package services

import (
	"context"
	"errors"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type fakeProductRepository struct {
	products []models.Product
	product  models.Product
	err      error
	filter   repositories.ProductFilter
	slug     string
}

func (repository *fakeProductRepository) FindAll(ctx context.Context, filter repositories.ProductFilter) ([]models.Product, error) {
	repository.filter = filter
	return repository.products, repository.err
}

func (repository *fakeProductRepository) FindBySlug(ctx context.Context, slug string) (models.Product, error) {
	repository.slug = slug
	return repository.product, repository.err
}

func TestProductServiceListProducts(t *testing.T) {
	featured := true
	repository := &fakeProductRepository{products: []models.Product{{ID: "product-1"}}}
	service := NewProductService(repository)

	products, err := service.ListProducts(context.Background(), dtos.ProductListQuery{
		Category:   " life ",
		IsFeatured: &featured,
		Limit:      3,
	})
	if err != nil {
		t.Fatalf("ListProducts() error = %v", err)
	}
	if len(products) != 1 || products[0].ID != "product-1" {
		t.Fatalf("ListProducts() = %+v, want product-1", products)
	}
	if repository.filter.Category != "life" || repository.filter.IsFeatured == nil || !*repository.filter.IsFeatured || repository.filter.Limit != 3 {
		t.Fatalf("FindAll filter = %+v, want normalized query", repository.filter)
	}
}

func TestProductServiceGetProductBySlug(t *testing.T) {
	repository := &fakeProductRepository{product: productFixture()}
	service := NewProductService(repository)

	product, err := service.GetProductBySlug(context.Background(), "secure-life-plus")
	if err != nil {
		t.Fatalf("GetProductBySlug() error = %v", err)
	}
	if product.ID != "product-1" || repository.slug != "secure-life-plus" {
		t.Fatalf("GetProductBySlug() = %+v, slug = %q", product, repository.slug)
	}
}

func TestProductServiceCreateProductQuote(t *testing.T) {
	repository := &fakeProductRepository{product: productFixture()}
	service := NewProductService(repository)

	quote, err := service.CreateProductQuote(context.Background(), "secure-life-plus", quoteInputFixture())
	if err != nil {
		t.Fatalf("CreateProductQuote() error = %v", err)
	}
	if quote.ProductID != "product-1" {
		t.Fatalf("ProductID = %q, want product-1", quote.ProductID)
	}
	if quote.Currency != constants.CurrencyIDR {
		t.Fatalf("Currency = %q, want %q", quote.Currency, constants.CurrencyIDR)
	}
	if quote.EstimatedAnnualPremium != 1512000 {
		t.Fatalf("EstimatedAnnualPremium = %d, want 1512000", quote.EstimatedAnnualPremium)
	}
	if quote.EstimatedPremium != 139000 {
		t.Fatalf("EstimatedPremium = %d, want 139000", quote.EstimatedPremium)
	}
	if len(quote.Notes) == 0 {
		t.Fatal("Notes is empty")
	}
}

func TestProductServiceCreateProductQuoteReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("db failed")
	service := NewProductService(&fakeProductRepository{err: expectedErr})

	_, err := service.CreateProductQuote(context.Background(), "secure-life-plus", quoteInputFixture())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("CreateProductQuote() error = %v, want %v", err, expectedErr)
	}
}

func TestProductServiceCreateProductQuoteValidatesProductLimits(t *testing.T) {
	tests := []struct {
		name  string
		input dtos.CreateProductQuoteInput
		want  error
	}{
		{name: "sum assured too low", input: withQuoteInput(func(input *dtos.CreateProductQuoteInput) { input.SumAssured = 99_000_000 }), want: constants.QuoteSumAssuredOutOfRangeError},
		{name: "sum assured too high", input: withQuoteInput(func(input *dtos.CreateProductQuoteInput) { input.SumAssured = 2_000_000_001 }), want: constants.QuoteSumAssuredOutOfRangeError},
		{name: "payment term too low", input: withQuoteInput(func(input *dtos.CreateProductQuoteInput) { input.PaymentTerm = 4 }), want: constants.QuotePaymentTermOutOfRangeError},
		{name: "payment term too high", input: withQuoteInput(func(input *dtos.CreateProductQuoteInput) { input.PaymentTerm = 31 }), want: constants.QuotePaymentTermOutOfRangeError},
	}

	service := NewProductService(&fakeProductRepository{product: productFixture()})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateProductQuote(context.Background(), "secure-life-plus", tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("CreateProductQuote() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestProductServiceCreateProductQuoteRequiresCompletePricingRules(t *testing.T) {
	product := productFixture()
	product.PricingRules.GenderFactors = nil
	service := NewProductService(&fakeProductRepository{product: product})

	_, err := service.CreateProductQuote(context.Background(), "secure-life-plus", quoteInputFixture())
	if !errors.Is(err, constants.QuotePricingRulesInvalidError) {
		t.Fatalf("CreateProductQuote() error = %v, want %v", err, constants.QuotePricingRulesInvalidError)
	}
}

func TestQuoteHelpers(t *testing.T) {
	if got := calculateTermFactor(5, 10); got != 1.05 {
		t.Fatalf("calculateTermFactor() = %v, want 1.05", got)
	}
	if got := calculateTermFactor(10, 5); got != 1 {
		t.Fatalf("calculateTermFactor() = %v, want 1", got)
	}
	if got := findAgeFactor([]models.AgeFactor{{MinAge: 18, MaxAge: 30, Factor: 1.1}}, 40); got != 0 {
		t.Fatalf("findAgeFactor() = %v, want 0", got)
	}
	if got := findRuleFactor(nil, "missing"); got != 0 {
		t.Fatalf("findRuleFactor() = %v, want 0", got)
	}
	if !validFactor(1.2) || validFactor(0) {
		t.Fatal("validFactor() returned unexpected result")
	}
	if divisor, ok := paymentFrequencyDivisor(constants.PaymentFrequencyMonthly); !ok || divisor != 12 {
		t.Fatalf("paymentFrequencyDivisor(monthly) = %v, %v; want 12, true", divisor, ok)
	}
	if divisor, ok := paymentFrequencyDivisor("weekly"); ok || divisor != 0 {
		t.Fatalf("paymentFrequencyDivisor(weekly) = %v, %v; want 0, false", divisor, ok)
	}
	if got := roundUpToNearestThousand(1001); got != 2000 {
		t.Fatalf("roundUpToNearestThousand() = %d, want 2000", got)
	}
}

func productFixture() models.Product {
	return models.Product{
		ID:              "product-1",
		Name:            "Secure Life Plus",
		Slug:            "secure-life-plus",
		Category:        models.ProductCategoryLife,
		MinSumAssured:   100_000_000,
		MaxSumAssured:   2_000_000_000,
		MinPaymentTerm:  5,
		MaxPaymentTerm:  30,
		StartingPremium: 100_000,
		PricingRules: models.PricingRules{
			BaseRate: 0.004,
			AgeFactors: []models.AgeFactor{
				{MinAge: 18, MaxAge: 30, Factor: 1.0},
				{MinAge: 31, MaxAge: 45, Factor: 1.2},
			},
			GenderFactors: map[string]float64{
				constants.GenderMale:   1.0,
				constants.GenderFemale: 0.95,
			},
			SmokerFactors: map[string]float64{
				constants.SmokerNo:  1.0,
				constants.SmokerYes: 1.4,
			},
			OccupationFactors: map[string]float64{
				constants.OccupationLow:      0.95,
				constants.OccupationStandard: 1.0,
				constants.OccupationHigh:     1.25,
			},
			HealthFactors: map[string]float64{
				constants.HealthRiskLow:    1.0,
				constants.HealthRiskMedium: 1.15,
				constants.HealthRiskHigh:   1.5,
			},
			FrequencyLoading: map[string]float64{
				constants.PaymentFrequencyAnnual:     1.0,
				constants.PaymentFrequencySemiAnnual: 1.03,
				constants.PaymentFrequencyQuarterly:  1.06,
				constants.PaymentFrequencyMonthly:    1.10,
			},
		},
	}
}

func quoteInputFixture() dtos.CreateProductQuoteInput {
	return dtos.CreateProductQuoteInput{
		Age:              35,
		Gender:           constants.GenderMale,
		SumAssured:       300_000_000,
		PaymentTerm:      10,
		PaymentFrequency: constants.PaymentFrequencyMonthly,
		Smoker:           constants.SmokerNo,
		OccupationClass:  constants.OccupationStandard,
		HealthRisk:       constants.HealthRiskLow,
	}
}

func withQuoteInput(update func(*dtos.CreateProductQuoteInput)) dtos.CreateProductQuoteInput {
	input := quoteInputFixture()
	update(&input)
	return input
}
