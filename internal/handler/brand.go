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

type brandService interface {
	Create(
		context.Context,
		service.CreateBrandInput,
	) (domain.Brand, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Brand, error)

	List(
		context.Context,
	) ([]domain.Brand, error)

	Update(
		context.Context,
		uuid.UUID,
		service.UpdateBrandInput,
	) (domain.Brand, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.Brand, error)
}

type BrandHandler struct {
	service brandService
	log     zerolog.Logger
}

func NewBrandHandler(
	service brandService,
	log zerolog.Logger,
) *BrandHandler {
	return &BrandHandler{
		service: service,
		log:     log,
	}
}

type createBrandRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type updateBrandRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type setBrandActiveRequest struct {
	IsActive *bool `json:"is_active"`
}

type brandResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *BrandHandler) Create(c *gin.Context) {
	var request createBrandRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	brand, err := h.service.Create(
		c.Request.Context(),
		service.CreateBrandInput{
			Name: request.Name,
			Code: request.Code,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create brand",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toBrandResponse(brand),
	)
}

func (h *BrandHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid brand id"},
		)
		return
	}

	brand, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get brand",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toBrandResponse(brand),
	)
}

func (h *BrandHandler) List(c *gin.Context) {
	brands, err := h.service.List(
		c.Request.Context(),
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list brands",
			err,
		)
		return
	}

	response := make(
		[]brandResponse,
		0,
		len(brands),
	)

	for _, brand := range brands {
		response = append(
			response,
			toBrandResponse(brand),
		)
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

func (h *BrandHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid brand id"},
		)
		return
	}

	var request updateBrandRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	brand, err := h.service.Update(
		c.Request.Context(),
		id,
		service.UpdateBrandInput{
			Name: request.Name,
			Code: request.Code,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update brand",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toBrandResponse(brand),
	)
}

func (h *BrandHandler) SetActive(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid brand id"},
		)
		return
	}

	var request setBrandActiveRequest

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

	brand, err := h.service.SetActive(
		c.Request.Context(),
		id,
		*request.IsActive,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"set brand active",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toBrandResponse(brand),
	)
}

func toBrandResponse(
	brand domain.Brand,
) brandResponse {
	return brandResponse{
		ID:        brand.ID,
		Name:      brand.Name,
		Code:      brand.Code,
		IsActive:  brand.IsActive,
		CreatedAt: brand.CreatedAt,
		UpdatedAt: brand.UpdatedAt,
	}
}
