package controllers

import (
	"errors"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
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
	statusQuery := ctx.Query("status")
	var statusArgs []string
	if statusQuery != "" {
		statusArgs = append(statusArgs, statusQuery)
	}

	query, err := validations.ValidateProductListQuery(
		ctx.Query(constants.ProductQueryCategory),
		ctx.Query(constants.ProductQueryFeatured),
		ctx.Query(constants.ProductQueryLimit),
		ctx.Query(constants.ProductQuerySearch),
		statusArgs...,
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

	quoteInput := dtos.ProductQuoteRequestToInput(request)
	quote, err := controller.productService.CreateProductQuote(ctx.Context(), slug, quoteInput)
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

func (controller *ProductController) GetPricingRules(ctx *fiber.Ctx) error {
	slug, err := validations.ValidateProductSlug(ctx.Params("slug"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	rules, err := controller.productService.GetPricingRules(ctx.Context(), slug)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve product pricing rules",
		})
	}

	return ctx.JSON(fiber.Map{
		"data": rules,
	})
}

func (controller *ProductController) GetByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product id is required",
		})
	}

	product, err := controller.productService.GetProductByID(ctx.Context(), id)
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

func (controller *ProductController) Create(ctx *fiber.Ctx) error {
	var request dtos.CreateProductRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product request body",
		})
	}

	product, err := controller.productService.CreateProduct(ctx.Context(), request)
	if err != nil {
		if errors.Is(err, constants.ErrProductSlugAlreadyExistsError) {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": product,
	})
}

func (controller *ProductController) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product id is required",
		})
	}

	var request dtos.UpdateProductRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product request body",
		})
	}

	product, err := controller.productService.UpdateProduct(ctx.Context(), id, request)
	if err != nil {
		if errors.Is(err, constants.ErrProductNotFoundError) || errors.Is(err, repositories.ErrProductNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrProductNotFound,
			})
		}
		if errors.Is(err, constants.ErrProductSlugAlreadyExistsError) {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"data": product,
	})
}

func (controller *ProductController) UpdateStatus(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product id is required",
		})
	}

	var request dtos.UpdateProductStatusRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	status, err := validations.ValidateProductStatus(request.Status)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	product, err := controller.productService.UpdateProductStatus(ctx.Context(), id, status)
	if err != nil {
		if errors.Is(err, constants.ErrProductNotFoundError) || errors.Is(err, repositories.ErrProductNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrProductNotFound,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": constants.ErrProductUpdateFailed,
		})
	}

	return ctx.JSON(fiber.Map{
		"data": product,
	})
}

func (controller *ProductController) ToggleStatus(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product id is required",
		})
	}

	product, err := controller.productService.ToggleProductStatus(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, constants.ErrProductNotFoundError) || errors.Is(err, repositories.ErrProductNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrProductNotFound,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": constants.ErrProductUpdateFailed,
		})
	}

	return ctx.JSON(fiber.Map{
		"data": product,
	})
}

func (controller *ProductController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product id is required",
		})
	}

	err := controller.productService.DeleteProduct(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, constants.ErrProductNotFoundError) || errors.Is(err, repositories.ErrProductNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrProductNotFound,
			})
		}
		if errors.Is(err, constants.ErrProductHasApplicationsError) {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": constants.ErrProductDeleteFailed,
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "product deleted successfully",
	})
}

func (controller *ProductController) GetMetrics(ctx *fiber.Ctx) error {
	metrics, err := controller.productService.GetProductManagementMetrics(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": constants.ErrProductMetricsFailed,
		})
	}

	return ctx.JSON(fiber.Map{
		"data": metrics,
	})
}

func (controller *ProductController) UpdatePricingRules(ctx *fiber.Ctx) error {
	slug, err := validations.ValidateProductSlug(ctx.Params("slug"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var rules []models.ProductPricingRule
	if err := ctx.BodyParser(&rules); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid pricing rules body",
		})
	}

	err = controller.productService.UpdatePricingRules(ctx.Context(), slug, rules)
	if err != nil {
		if errors.Is(err, constants.ErrProductNotFoundError) || errors.Is(err, repositories.ErrProductNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrProductNotFound,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": constants.ErrProductPricingRulesUpdateFailed,
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "pricing rules updated successfully",
	})
}
