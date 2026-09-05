package repositories

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPostgresProductRepository(t *testing.T) {
	db := sqliteDB(t)
	seedProduct(t, db, models.Product{ID: "product-1", Name: "Secure Life Plus", Slug: "secure-life-plus", Category: models.ProductCategoryLife, IsFeatured: true})
	seedProduct(t, db, models.Product{ID: "product-2", Name: "Health Guard", Slug: "health-guard", Category: models.ProductCategoryHealth, IsFeatured: false})

	repository := NewPostgresProductRepository(db)
	featured := true
	products, err := repository.FindAll(context.Background(), ProductFilter{Category: "life", IsFeatured: &featured, Limit: 1})
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(products) != 1 || products[0].ID != "product-1" {
		t.Fatalf("FindAll() = %+v, want product-1", products)
	}

	product, err := repository.FindBySlug(context.Background(), "secure-life-plus")
	if err != nil {
		t.Fatalf("FindBySlug() error = %v", err)
	}
	if product.ID != "product-1" {
		t.Fatalf("FindBySlug() = %+v, want product-1", product)
	}

	_, err = repository.FindBySlug(context.Background(), "missing")
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("FindBySlug(missing) error = %v, want ErrProductNotFound", err)
	}
}

func TestPostgresApplicationRepository(t *testing.T) {
	db := sqliteDB(t)
	product := models.Product{ID: "product-1", Name: "Secure Life Plus", Slug: "secure-life-plus", Category: models.ProductCategoryLife, IsFeatured: true}
	seedProduct(t, db, product)

	repository := NewPostgresApplicationRepository(db)
	application := models.Application{ID: "application-1", ProductID: product.ID, FullName: "Bayu", Email: "bayu@example.com", Phone: "+628123456789", Age: 35, Gender: "male", SumAssured: 300000000, PaymentTerm: 10, PaymentFrequency: "monthly", Smoker: "no", OccupationClass: "standard", HealthRisk: "low", Premium: 139000, Status: models.ApplicationStatusSubmitted, ReviewChecks: []models.ApplicationReviewCheck{{ID: "check-1", ApplicationID: "application-1", CheckType: models.ApplicationReviewCheckTypeIdentityVerified, Status: models.ApplicationReviewCheckStatusPending}}}
	if err := repository.Create(context.Background(), &application); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repository.FindByID(context.Background(), application.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.ID != application.ID || found.Product.ID != product.ID || len(found.ReviewChecks) != 1 {
		t.Fatalf("FindByID() = %+v, want preloaded product and checks", found)
	}

	reviewedAt := time.Now().UTC()
	if err := repository.UpdateStatus(context.Background(), application.ID, models.ApplicationStatusUnderReview, "underwriter", "", reviewedAt); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	found, err = repository.FindByID(context.Background(), application.ID)
	if err != nil {
		t.Fatalf("FindByID() after update error = %v", err)
	}
	if found.Status != models.ApplicationStatusUnderReview || found.ReviewedBy != "underwriter" || found.ReviewedAt == nil {
		t.Fatalf("updated application = %+v, want reviewed status", found)
	}

	_, err = repository.FindByID(context.Background(), "missing")
	if !errors.Is(err, ErrApplicationNotFound) {
		t.Fatalf("FindByID(missing) error = %v, want ErrApplicationNotFound", err)
	}
	if err := repository.UpdateStatus(context.Background(), "missing", models.ApplicationStatusApproved, "underwriter", "", reviewedAt); !errors.Is(err, ErrApplicationNotFound) {
		t.Fatalf("UpdateStatus(missing) error = %v, want ErrApplicationNotFound", err)
	}

	applications, total, err := repository.List(context.Background(), ApplicationListFilter{ProductID: product.ID, Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 1 || len(applications) != 1 {
		t.Fatalf("List() = %+v total=%d, want 1 application", applications, total)
	}
}

func TestPostgresApplicationReviewCheckRepository(t *testing.T) {
	db := sqliteDB(t)
	product := models.Product{ID: "product-1", Name: "Secure Life Plus", Slug: "secure-life-plus", Category: models.ProductCategoryLife, IsFeatured: true}
	seedProduct(t, db, product)
	application := models.Application{ID: "application-1", ProductID: product.ID, FullName: "Bayu", Email: "bayu@example.com", Phone: "+628123456789", Age: 35, Gender: "male", SumAssured: 300000000, PaymentTerm: 10, PaymentFrequency: "monthly", Smoker: "no", OccupationClass: "standard", HealthRisk: "low", Premium: 139000, Status: models.ApplicationStatusSubmitted}
	if err := db.Create(&application).Error; err != nil {
		t.Fatalf("seed application error = %v", err)
	}
	check := models.ApplicationReviewCheck{ID: "check-1", ApplicationID: application.ID, CheckType: models.ApplicationReviewCheckTypeIdentityVerified, Status: models.ApplicationReviewCheckStatusPending}
	if err := db.Create(&check).Error; err != nil {
		t.Fatalf("seed check error = %v", err)
	}

	repository := NewPostgresApplicationReviewCheckRepository(db)
	checks, err := repository.FindByApplicationID(context.Background(), application.ID)
	if err != nil {
		t.Fatalf("FindByApplicationID() error = %v", err)
	}
	if len(checks) != 1 || checks[0].ID != check.ID {
		t.Fatalf("FindByApplicationID() = %+v, want check-1", checks)
	}

	reviewedAt := time.Now().UTC()
	if err := repository.UpdateStatus(context.Background(), application.ID, models.ApplicationReviewCheckTypeIdentityVerified, models.ApplicationReviewCheckStatusPassed, "underwriter", "ok", reviewedAt); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	checks, err = repository.FindByApplicationID(context.Background(), application.ID)
	if err != nil {
		t.Fatalf("FindByApplicationID() after update error = %v", err)
	}
	if checks[0].Status != models.ApplicationReviewCheckStatusPassed || checks[0].ReviewedBy != "underwriter" || checks[0].Notes != "ok" || checks[0].ReviewedAt == nil {
		t.Fatalf("updated check = %+v, want reviewed check", checks[0])
	}

	err = repository.UpdateStatus(context.Background(), application.ID, models.ApplicationReviewCheckTypeIncomeVerified, models.ApplicationReviewCheckStatusPassed, "underwriter", "ok", reviewedAt)
	if !errors.Is(err, ErrApplicationReviewCheckNotFound) {
		t.Fatalf("UpdateStatus(missing) error = %v, want ErrApplicationReviewCheckNotFound", err)
	}
}

func sqliteDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&models.Product{}, &models.Application{}, &models.ApplicationReviewCheck{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}

func seedProduct(t *testing.T, db *gorm.DB, product models.Product) {
	t.Helper()
	product.ShortDescription = "short"
	product.Description = "description"
	product.TargetCustomer = "customer"
	product.MinSumAssured = 100000000
	product.MaxSumAssured = 2000000000
	product.MinPaymentTerm = 5
	product.MaxPaymentTerm = 30
	product.StartingPremium = 100000
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("seed product error = %v", err)
	}
}
