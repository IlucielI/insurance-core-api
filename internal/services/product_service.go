package services

import (
	"context"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type ProductService struct {
	productRepository repositories.ProductRepository
}

type ListProductsInput struct {
	Category   string
	IsFeatured *bool
	Limit      int
}

func NewProductService(productRepository repositories.ProductRepository) *ProductService {
	return &ProductService{productRepository: productRepository}
}

func (service *ProductService) ListProducts(ctx context.Context, input ListProductsInput) ([]models.Product, error) {
	return service.productRepository.FindAll(ctx, repositories.ProductFilter{
		Category:   strings.TrimSpace(input.Category),
		IsFeatured: input.IsFeatured,
		Limit:      input.Limit,
	})
}

func (service *ProductService) GetProductBySlug(ctx context.Context, slug string) (models.Product, error) {
	return service.productRepository.FindBySlug(ctx, slug)
}
