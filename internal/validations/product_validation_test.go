package validations

import (
	"strings"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

func TestValidateProductListQuery(t *testing.T) {
	query, err := ValidateProductListQuery(" life ", "true", "10", " term ")
	if err != nil {
		t.Fatalf("ValidateProductListQuery() error = %v", err)
	}
	if query.Category != "life" || query.IsFeatured == nil || !*query.IsFeatured || query.Limit != 10 || query.Search != "term" {
		t.Fatalf("ValidateProductListQuery() = %+v, want normalized values", query)
	}

	tests := []struct {
		name     string
		category string
		featured string
		limit    string
		search   string
		wantErr  string
	}{
		{name: "invalid category", category: "travel", wantErr: constants.ErrProductCategoryInvalid},
		{name: "invalid featured", featured: "yes", wantErr: constants.ErrProductFeaturedInvalid},
		{name: "invalid limit", limit: "0", wantErr: constants.ErrProductLimitInvalid},
		{name: "limit too high", limit: "51", wantErr: constants.ErrProductLimitTooHigh},
		{name: "search too long", search: strings.Repeat("a", 101), wantErr: constants.ErrProductSearchInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateProductListQuery(tt.category, tt.featured, tt.limit, tt.search)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("ValidateProductListQuery() error = %v, want %q", err, tt.wantErr)
			}
		})
	}

	// Test default status is active
	defaultQuery, err := ValidateProductListQuery("", "", "", "")
	if err != nil {
		t.Fatalf("ValidateProductListQuery() unexpected error = %v", err)
	}
	if defaultQuery.Status != string(models.ProductStatusActive) {
		t.Fatalf("defaultQuery.Status = %q, want active", defaultQuery.Status)
	}

	// Test explicit status all
	allQuery, err := ValidateProductListQuery("", "", "", "", "all")
	if err != nil || allQuery.Status != "all" {
		t.Fatalf("allQuery.Status = %q, want all", allQuery.Status)
	}

	// Test explicit status draft
	draftQuery, err := ValidateProductListQuery("", "", "", "", "draft")
	if err != nil || draftQuery.Status != "draft" {
		t.Fatalf("draftQuery.Status = %q, want draft", draftQuery.Status)
	}
}

func TestValidateProductSlug(t *testing.T) {
	slug, err := ValidateProductSlug(" secure-life-plus ")
	if err != nil {
		t.Fatalf("ValidateProductSlug() error = %v", err)
	}
	if slug != "secure-life-plus" {
		t.Fatalf("ValidateProductSlug() = %q, want secure-life-plus", slug)
	}

	_, err = ValidateProductSlug(" ")
	if err == nil || err.Error() != constants.ErrProductSlugRequired {
		t.Fatalf("ValidateProductSlug() error = %v, want slug required", err)
	}
}

func TestValidateProductQuoteRequest(t *testing.T) {
	request, err := ValidateProductQuoteRequest(productQuoteRequestFixture())
	if err != nil {
		t.Fatalf("ValidateProductQuoteRequest() error = %v", err)
	}
	if request.Gender != constants.GenderMale || request.PaymentFrequency != constants.PaymentFrequencyMonthly {
		t.Fatalf("ValidateProductQuoteRequest() = %+v, want normalized request", request)
	}

	tests := []struct {
		name    string
		update  func(*dtos.ProductQuoteRequest)
		wantErr string
	}{
		{name: "age", update: func(request *dtos.ProductQuoteRequest) { request.Age = 17 }, wantErr: constants.ErrQuoteAgeInvalid},
		{name: "gender", update: func(request *dtos.ProductQuoteRequest) { request.Gender = "" }, wantErr: constants.ErrQuoteGenderInvalid},
		{name: "sum assured", update: func(request *dtos.ProductQuoteRequest) { request.SumAssured = 0 }, wantErr: constants.ErrQuoteSumAssuredInvalid},
		{name: "payment term", update: func(request *dtos.ProductQuoteRequest) { request.PaymentTerm = 0 }, wantErr: constants.ErrQuotePaymentTermInvalid},
		{name: "frequency", update: func(request *dtos.ProductQuoteRequest) { request.PaymentFrequency = "weekly" }, wantErr: constants.ErrQuotePaymentFrequencyInvalid},
		{name: "smoker", update: func(request *dtos.ProductQuoteRequest) { request.Smoker = "sometimes" }, wantErr: constants.ErrQuoteSmokerInvalid},
		{name: "occupation", update: func(request *dtos.ProductQuoteRequest) { request.OccupationClass = "danger" }, wantErr: constants.ErrQuoteOccupationInvalid},
		{name: "health", update: func(request *dtos.ProductQuoteRequest) { request.HealthRisk = "critical" }, wantErr: constants.ErrQuoteHealthRiskInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := productQuoteRequestFixture()
			tt.update(&request)
			_, err := ValidateProductQuoteRequest(request)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("ValidateProductQuoteRequest() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProductQuoteRequestDefaultsAndNormalization(t *testing.T) {
	req := dtos.ProductQuoteRequest{
		Age:              25,
		Gender:           "FEMALE",
		SumAssured:       100_000_000,
		PaymentTerm:      10,
		PaymentFrequency: "annually",
		Smoker:           "NO",
		OccupationClass:  "",
		HealthRisk:       "",
	}
	validated, err := ValidateProductQuoteRequest(req)
	if err != nil {
		t.Fatalf("ValidateProductQuoteRequest() error = %v", err)
	}
	if validated.Gender != constants.GenderFemale {
		t.Fatalf("Gender = %q, want %q", validated.Gender, constants.GenderFemale)
	}
	if validated.PaymentFrequency != constants.PaymentFrequencyAnnual {
		t.Fatalf("PaymentFrequency = %q, want %q", validated.PaymentFrequency, constants.PaymentFrequencyAnnual)
	}
	if validated.Smoker != constants.SmokerNo {
		t.Fatalf("Smoker = %q, want %q", validated.Smoker, constants.SmokerNo)
	}
	if validated.OccupationClass != constants.OccupationStandard {
		t.Fatalf("OccupationClass = %q, want %q", validated.OccupationClass, constants.OccupationStandard)
	}
	if validated.HealthRisk != constants.HealthRiskLow {
		t.Fatalf("HealthRisk = %q, want %q", validated.HealthRisk, constants.HealthRiskLow)
	}
}

func TestValidateApplicationRequest(t *testing.T) {
	request := dtos.CreateApplicationRequest{
		FullName:            " Bayu Anugerah ",
		Email:               " bayu@example.com ",
		Phone:               " +628123456789 ",
		ProductQuoteRequest: productQuoteRequestFixture(),
	}

	validated, err := ValidateApplicationRequest(request)
	if err != nil {
		t.Fatalf("ValidateApplicationRequest() error = %v", err)
	}
	if validated.FullName != "Bayu Anugerah" || validated.Email != "bayu@example.com" || validated.Phone != "+628123456789" {
		t.Fatalf("ValidateApplicationRequest() = %+v, want trimmed applicant fields", validated)
	}

	tests := []struct {
		name    string
		update  func(*dtos.CreateApplicationRequest)
		wantErr string
	}{
		{name: "full name", update: func(request *dtos.CreateApplicationRequest) { request.FullName = "x" }, wantErr: constants.ErrApplicationFullNameInvalid},
		{name: "email", update: func(request *dtos.CreateApplicationRequest) { request.Email = "invalid" }, wantErr: constants.ErrApplicationEmailInvalid},
		{name: "phone", update: func(request *dtos.CreateApplicationRequest) { request.Phone = "123" }, wantErr: constants.ErrApplicationPhoneInvalid},
		{name: "quote", update: func(request *dtos.CreateApplicationRequest) { request.Age = 61 }, wantErr: constants.ErrQuoteAgeInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := dtos.CreateApplicationRequest{FullName: "Bayu Anugerah", Email: "bayu@example.com", Phone: "+628123456789", ProductQuoteRequest: productQuoteRequestFixture()}
			tt.update(&request)
			_, err := ValidateApplicationRequest(request)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("ValidateApplicationRequest() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateApplicationReviewCheckRequest(t *testing.T) {
	request, err := ValidateApplicationReviewCheckRequest(dtos.UpdateApplicationReviewCheckRequest{
		Status:     " passed ",
		ReviewedBy: " underwriter ",
		Notes:      " ok ",
	})
	if err != nil {
		t.Fatalf("ValidateApplicationReviewCheckRequest() error = %v", err)
	}
	if request.Status != "passed" || request.ReviewedBy != "underwriter" || request.Notes != "ok" {
		t.Fatalf("ValidateApplicationReviewCheckRequest() = %+v, want trimmed request", request)
	}

	tests := []dtos.UpdateApplicationReviewCheckRequest{
		{Status: "unknown", ReviewedBy: "underwriter"},
		{Status: "passed", ReviewedBy: ""},
		{Status: "passed", ReviewedBy: "underwriter", Notes: string(make([]byte, 501))},
	}
	for _, tt := range tests {
		if _, err := ValidateApplicationReviewCheckRequest(tt); err == nil {
			t.Fatalf("ValidateApplicationReviewCheckRequest(%+v) error = nil, want error", tt)
		}
	}
}

func TestValidateApplicationReviewCheckType(t *testing.T) {
	checkType, err := ValidateApplicationReviewCheckType(" identity_verified ")
	if err != nil {
		t.Fatalf("ValidateApplicationReviewCheckType() error = %v", err)
	}
	if checkType != "identity_verified" {
		t.Fatalf("ValidateApplicationReviewCheckType() = %q, want identity_verified", checkType)
	}

	if _, err := ValidateApplicationReviewCheckType("unknown"); err == nil {
		t.Fatal("ValidateApplicationReviewCheckType(unknown) error = nil, want error")
	}
}

func TestContains(t *testing.T) {
	if !contains("male", constants.GenderMale, constants.GenderFemale) {
		t.Fatal("contains() = false, want true")
	}
	if contains("", constants.GenderMale, constants.GenderFemale) {
		t.Fatal("contains(empty) = true, want false")
	}
	if contains("unknown", constants.GenderMale, constants.GenderFemale) {
		t.Fatal("contains(unknown) = true, want false")
	}
}

func productQuoteRequestFixture() dtos.ProductQuoteRequest {
	return dtos.ProductQuoteRequest{
		Age:              35,
		Gender:           " male ",
		SumAssured:       300_000_000,
		PaymentTerm:      10,
		PaymentFrequency: " monthly ",
		Smoker:           " no ",
		OccupationClass:  " standard ",
		HealthRisk:       " low ",
	}
}

func TestValidateApplicationListQuery(t *testing.T) {
	cases := []struct{ name, status, product, page, limit, want string }{
		{"defaults", "", "", "", "", ""},
		{"status", "submitted", " product-1 ", "2", "10", ""},
		{"status draft", "draft", "", "", "", ""},
		{"bad status", "unknown", "", "", "", constants.ErrApplicationStatusInvalid},
		{"bad product", "", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "", "", constants.ErrApplicationListFilterInvalid},
		{"bad page", "", "", "x", "", constants.ErrApplicationPageInvalid},
		{"bad limit", "", "", "", "0", constants.ErrApplicationLimitInvalid},
		{"large limit", "", "", "", "51", constants.ErrApplicationListLimitTooHigh},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := ValidateApplicationListQuery(tc.status, tc.product, tc.page, tc.limit)
			if tc.want != "" {
				if err == nil || err.Error() != tc.want {
					t.Fatalf("error=%v want=%s", err, tc.want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if q.Limit == 0 {
				t.Fatal("default limit missing")
			}
		})
	}
}

func TestValidateProductListQueryWithStatus(t *testing.T) {
	query, err := ValidateProductListQuery("life", "", "10", "", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query.Status != "active" {
		t.Fatalf("want status active, got %s", query.Status)
	}

	queryAll, err := ValidateProductListQuery("", "", "", "", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if queryAll.Status != "all" {
		t.Fatalf("want status all, got %s", queryAll.Status)
	}

	_, err = ValidateProductListQuery("", "", "", "", "invalid_status")
	if err == nil || err.Error() != constants.ErrProductStatusInvalid {
		t.Fatalf("want ErrProductStatusInvalid, got %v", err)
	}
}

func TestValidateCreateProductRequest(t *testing.T) {
	validReq := dtos.CreateProductRequest{
		Name:             "Proteksi Jiwa Utama",
		Slug:             "proteksi-jiwa-utama",
		Category:         "life",
		Status:           "active",
		ShortDescription: "Asuransi jiwa murni",
		Description:      "Deskripsi lengkap proteksi jiwa",
		TargetCustomer:   "Keluarga muda",
		MinSumAssured:    50000000,
		MaxSumAssured:    1000000000,
		MinPaymentTerm:   5,
		MaxPaymentTerm:   20,
		StartingPremium:  150000,
		NonMCULimit:      500000000,
		Benefits:         []string{" Santunan Meninggal Dunia ", " Santunan Cacat Tetap "},
	}

	res, err := ValidateCreateProductRequest(validReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Benefits[0] != "Santunan Meninggal Dunia" {
		t.Fatalf("expected benefits trimmed, got %q", res.Benefits[0])
	}

	// Test invalid slug
	invalidSlugReq := validReq
	invalidSlugReq.Slug = "invalid slug with spaces"
	_, err = ValidateCreateProductRequest(invalidSlugReq)
	if err == nil || err.Error() != constants.ErrProductSlugInvalid {
		t.Fatalf("want ErrProductSlugInvalid, got %v", err)
	}

	// Test min/max sum assured inverted
	invalidSumReq := validReq
	invalidSumReq.MinSumAssured = 1000000
	invalidSumReq.MaxSumAssured = 500000
	_, err = ValidateCreateProductRequest(invalidSumReq)
	if err == nil || err.Error() != constants.ErrProductMaxSumAssuredInvalid {
		t.Fatalf("want ErrProductMaxSumAssuredInvalid, got %v", err)
	}
}

func TestValidateUpdateProductRequest(t *testing.T) {
	name := "Updated Name"
	slug := "updated-name"
	minSum := int64(10000000)
	maxSum := int64(500000000)
	req := dtos.UpdateProductRequest{
		Name:          &name,
		Slug:          &slug,
		MinSumAssured: &minSum,
		MaxSumAssured: &maxSum,
	}
	valid, err := ValidateUpdateProductRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *valid.Name != "Updated Name" {
		t.Fatalf("expected name updated, got %s", *valid.Name)
	}
}

