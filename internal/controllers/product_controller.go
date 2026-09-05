package controllers

import (
	"errors"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
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

	products, err := controller.productService.ListProducts(ctx.Context(), query)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": constants.ErrProductListFailed,
		})
	}

	return ctx.JSON(fiber.Map{
		"data": products,
	})
}

func (controller *ProductController) CreateQuote(ctx *fiber.Ctx) error {
	slug, err := validations.ValidateProductSlug(ctx.Params("slug"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var request dtos.ProductQuoteRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": constants.ErrProductQuoteBodyInvalid,
		})
	}

	request, err = validations.ValidateProductQuoteRequest(request)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	quote, err := controller.productService.CreateProductQuote(ctx.Context(), slug, dtos.CreateProductQuoteInput{
		Age:              request.Age,
		Gender:           request.Gender,
		SumAssured:       request.SumAssured,
		PaymentTerm:      request.PaymentTerm,
		PaymentFrequency: request.PaymentFrequency,
		Smoker:           request.Smoker,
		OccupationClass:  request.OccupationClass,
		HealthRisk:       request.HealthRisk,
	})
	if err != nil {
		if errors.Is(err, repositories.ErrProductNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrProductNotFound,
			})
		}

		if constants.IsQuoteValidationError(err) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": constants.ErrProductQuoteFailed,
		})
	}

	return ctx.JSON(fiber.Map{
		"data": quote,
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
				"error": constants.ErrProductNotFound,
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
