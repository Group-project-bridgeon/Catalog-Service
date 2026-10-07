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

type productImageService interface {
	Create(
		context.Context,
		service.CreateProductImageInput,
	) (domain.ProductImage, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.ProductImage, error)

	List(
		context.Context,
	) ([]domain.ProductImage, error)

	ListByProductID(
		context.Context,
		uuid.UUID,
	) ([]domain.ProductImage, error)

	ListByVariantID(
		context.Context,
		uuid.UUID,
	) ([]domain.ProductImage, error)

	Update(
		context.Context,
		uuid.UUID,
		service.UpdateProductImageInput,
	) (domain.ProductImage, error)
}

type ProductImageHandler struct {
	service productImageService
	log     zerolog.Logger
}

func NewProductImageHandler(
	service productImageService,
	log zerolog.Logger,
) *ProductImageHandler {
	return &ProductImageHandler{
		service: service,
		log:     log,
	}
}

type createProductImageRequest struct {
	ProductID uuid.UUID  `json:"product_id"`
	VariantID *uuid.UUID `json:"variant_id"`
	ImageURL  string     `json:"image_url"`
	SortOrder int32      `json:"sort_order"`
}

type updateProductImageRequest struct {
	VariantID *uuid.UUID `json:"variant_id"`
	ImageURL  string     `json:"image_url"`
	SortOrder int32      `json:"sort_order"`
}

type productImageResponse struct {
	ID        uuid.UUID  `json:"id"`
	ProductID uuid.UUID  `json:"product_id"`
	VariantID *uuid.UUID `json:"variant_id,omitempty"`
	ImageURL  string     `json:"image_url"`
	SortOrder int32      `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
}

func (h *ProductImageHandler) Create(
	c *gin.Context,
) {
	var request createProductImageRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	image, err := h.service.Create(
		c.Request.Context(),
		service.CreateProductImageInput{
			ProductID: request.ProductID,
			VariantID: request.VariantID,
			ImageURL:  request.ImageURL,
			SortOrder: request.SortOrder,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create product image",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toProductImageResponse(image),
	)
}

func (h *ProductImageHandler) GetByID(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product image id"},
		)
		return
	}

	image, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get product image",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductImageResponse(image),
	)
}

func (h *ProductImageHandler) List(
	c *gin.Context,
) {
	images, err := h.service.List(
		c.Request.Context(),
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list product images",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductImageResponses(images),
	)
}

func (h *ProductImageHandler) ListByProductID(
	c *gin.Context,
) {
	productID, err := uuid.Parse(
		c.Param("id"),
	)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product id"},
		)
		return
	}

	images, err := h.service.ListByProductID(
		c.Request.Context(),
		productID,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list product images by product",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductImageResponses(images),
	)
}

func (h *ProductImageHandler) ListByVariantID(
	c *gin.Context,
) {
	variantID, err := uuid.Parse(
		c.Param("id"),
	)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid variant id"},
		)
		return
	}

	images, err := h.service.ListByVariantID(
		c.Request.Context(),
		variantID,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list product images by variant",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductImageResponses(images),
	)
}

func (h *ProductImageHandler) Update(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product image id"},
		)
		return
	}

	var request updateProductImageRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	image, err := h.service.Update(
		c.Request.Context(),
		id,
		service.UpdateProductImageInput{
			VariantID: request.VariantID,
			ImageURL:  request.ImageURL,
			SortOrder: request.SortOrder,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update product image",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductImageResponse(image),
	)
}

func toProductImageResponse(
	image domain.ProductImage,
) productImageResponse {
	return productImageResponse{
		ID:        image.ID,
		ProductID: image.ProductID,
		VariantID: image.VariantID,
		ImageURL:  image.ImageURL,
		SortOrder: image.SortOrder,
		CreatedAt: image.CreatedAt,
	}
}

func toProductImageResponses(
	images []domain.ProductImage,
) []productImageResponse {
	response := make(
		[]productImageResponse,
		0,
		len(images),
	)

	for _, image := range images {
		response = append(
			response,
			toProductImageResponse(image),
		)
	}

	return response
}
