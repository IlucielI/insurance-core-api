package controllers

import (
	"errors"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
	"github.com/gofiber/fiber/v2"
)

type ProductController struct {
	productService *services.ProductService
}

func NewProductController(productService *services.ProductService) *ProductController {
	return &ProductController{productService: productService}
}

func (controller *ProductController) List(ctx *fiber.Ctx) error {
	query, err := validations.ValidateProductListQuery(
		ctx.Query(constants.ProductQueryCategory),
		ctx.Query(constants.ProductQueryFeatured),
		ctx.Query(constants.ProductQueryLimit),
	)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	products, err := controller.productService.ListProducts(ctx.Context(), services.ListProductsInput{
		Category:   query.Category,
		IsFeatured: query.IsFeatured,
		Limit:      query.Limit,
	})
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": constants.ErrProductListFailed,
		})
	}

	return ctx.JSON(fiber.Map{
		"data": products,
	})
}

func (controller *ProductController) Detail(ctx *fiber.Ctx) error {
	slug, err := validations.ValidateProductSlug(ctx.Params("slug"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	product, err := controller.productService.GetProductBySlug(ctx.Context(), slug)
	if err != nil {
		if errors.Is(err, repositories.ErrProductNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": constants.ErrProductDetailFailed,
		})
	}

	return ctx.JSON(fiber.Map{
		"data": product,
	})
}
