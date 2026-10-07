package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type productVariantService interface {
	Create(
		context.Context,
		service.CreateProductVariantInput,
	) (domain.ProductVariant, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.ProductVariant, error)

	List(
		context.Context,
	) ([]domain.ProductVariant, error)

	ListByProductID(
		context.Context,
		uuid.UUID,
	) ([]domain.ProductVariant, error)

	Update(
		context.Context,
		uuid.UUID,
		service.UpdateProductVariantInput,
	) (domain.ProductVariant, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.ProductVariant, error)
}

type ProductVariantHandler struct {
	service productVariantService
	log     zerolog.Logger
}

func NewProductVariantHandler(
	service productVariantService,
	log zerolog.Logger,
) *ProductVariantHandler {
	return &ProductVariantHandler{
		service: service,
		log:     log,
	}
}

type createProductVariantRequest struct {
	ProductID           uuid.UUID `json:"product_id"`
	SKU                 string    `json:"sku"`
	DefaultSellingPrice *string   `json:"default_selling_price"`
	UnitID              uuid.UUID `json:"unit_id"`
}

type updateProductVariantRequest struct {
	SKU                 string  `json:"sku"`
	DefaultSellingPrice *string `json:"default_selling_price"`
}

type setProductVariantActiveRequest struct {
	IsActive *bool `json:"is_active"`
}

type productVariantResponse struct {
	ID                  uuid.UUID `json:"id"`
	ProductID           uuid.UUID `json:"product_id"`
	SKU                 string    `json:"sku"`
	DefaultSellingPrice *string   `json:"default_selling_price,omitempty"`
	UnitID              uuid.UUID `json:"unit_id"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (h *ProductVariantHandler) Create(c *gin.Context) {
	var request createProductVariantRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	variant, err := h.service.Create(
		c.Request.Context(),
		service.CreateProductVariantInput{
			ProductID:           request.ProductID,
			SKU:                 request.SKU,
			DefaultSellingPrice: request.DefaultSellingPrice,
			UnitID:              request.UnitID,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create product variant",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toProductVariantResponse(variant),
	)
}

func (h *ProductVariantHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product variant id"},
		)
		return
	}

	variant, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get product variant",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductVariantResponse(variant),
	)
}

func (h *ProductVariantHandler) List(c *gin.Context) {
	variants, err := h.service.List(
		c.Request.Context(),
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list product variants",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductVariantResponses(variants),
	)
}

func (h *ProductVariantHandler) ListByProductID(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product id"},
		)
		return
	}

	variants, err := h.service.ListByProductID(
		c.Request.Context(),
		productID,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list product variants by product",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductVariantResponses(variants),
	)
}

func (h *ProductVariantHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product variant id"},
		)
		return
	}

	var request updateProductVariantRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	variant, err := h.service.Update(
		c.Request.Context(),
		id,
		service.UpdateProductVariantInput{
			SKU:                 request.SKU,
			DefaultSellingPrice: request.DefaultSellingPrice,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update product variant",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductVariantResponse(variant),
	)
}

func (h *ProductVariantHandler) SetActive(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product variant id"},
		)
		return
	}

	var request setProductVariantActiveRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	if request.IsActive == nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "is_active is required"},
		)
		return
	}

	variant, err := h.service.SetActive(
		c.Request.Context(),
		id,
		*request.IsActive,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"set product variant active",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductVariantResponse(variant),
	)
}

func toProductVariantResponse(
	variant domain.ProductVariant,
) productVariantResponse {
	return productVariantResponse{
		ID:                  variant.ID,
		ProductID:           variant.ProductID,
		SKU:                 variant.SKU,
		DefaultSellingPrice: variant.DefaultSellingPrice,
		UnitID:              variant.UnitID,
		IsActive:            variant.IsActive,
		CreatedAt:           variant.CreatedAt,
		UpdatedAt:           variant.UpdatedAt,
	}
}

func toProductVariantResponses(
	variants []domain.ProductVariant,
) []productVariantResponse {
	response := make(
		[]productVariantResponse,
		0,
		len(variants),
	)

	for _, variant := range variants {
		response = append(
			response,
			toProductVariantResponse(variant),
		)
	}

	return response
}
