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

type productService interface {
	Create(
		context.Context,
		service.CreateProductInput,
	) (domain.Product, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Product, error)

	List(
		context.Context,
	) ([]domain.Product, error)

	Update(
		context.Context,
		uuid.UUID,
		service.UpdateProductInput,
	) (domain.Product, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.Product, error)
}

type ProductHandler struct {
	service productService
	log     zerolog.Logger
}

func NewProductHandler(
	service productService,
	log zerolog.Logger,
) *ProductHandler {
	return &ProductHandler{
		service: service,
		log:     log,
	}
}

type createProductRequest struct {
	CategoryID  uuid.UUID  `json:"category_id"`
	BrandID     *uuid.UUID `json:"brand_id"`
	Name        string     `json:"name"`
	ProductCode string     `json:"product_code"`
	Description *string    `json:"description"`
}

type updateProductRequest struct {
	CategoryID  uuid.UUID  `json:"category_id"`
	BrandID     *uuid.UUID `json:"brand_id"`
	Name        string     `json:"name"`
	ProductCode string     `json:"product_code"`
	Description *string    `json:"description"`
}

type setProductActiveRequest struct {
	IsActive *bool `json:"is_active"`
}

type productResponse struct {
	ID          uuid.UUID  `json:"id"`
	CategoryID  uuid.UUID  `json:"category_id"`
	BrandID     *uuid.UUID `json:"brand_id,omitempty"`
	Name        string     `json:"name"`
	ProductCode string     `json:"product_code"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (h *ProductHandler) Create(c *gin.Context) {
	var request createProductRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	product, err := h.service.Create(
		c.Request.Context(),
		service.CreateProductInput{
			CategoryID:  request.CategoryID,
			BrandID:     request.BrandID,
			Name:        request.Name,
			ProductCode: request.ProductCode,
			Description: request.Description,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create product",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toProductResponse(product),
	)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product id"},
		)
		return
	}

	product, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get product",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductResponse(product),
	)
}

func (h *ProductHandler) List(c *gin.Context) {
	products, err := h.service.List(
		c.Request.Context(),
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list products",
			err,
		)
		return
	}

	response := make(
		[]productResponse,
		0,
		len(products),
	)

	for _, product := range products {
		response = append(
			response,
			toProductResponse(product),
		)
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product id"},
		)
		return
	}

	var request updateProductRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	product, err := h.service.Update(
		c.Request.Context(),
		id,
		service.UpdateProductInput{
			CategoryID:  request.CategoryID,
			BrandID:     request.BrandID,
			Name:        request.Name,
			ProductCode: request.ProductCode,
			Description: request.Description,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update product",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductResponse(product),
	)
}

func (h *ProductHandler) SetActive(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product id"},
		)
		return
	}

	var request setProductActiveRequest

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

	product, err := h.service.SetActive(
		c.Request.Context(),
		id,
		*request.IsActive,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"set product active",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toProductResponse(product),
	)
}

func toProductResponse(
	product domain.Product,
) productResponse {
	return productResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		BrandID:     product.BrandID,
		Name:        product.Name,
		ProductCode: product.ProductCode,
		Description: product.Description,
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}
