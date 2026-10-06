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

type categoryService interface {
	Create(
		context.Context,
		service.CreateCategoryInput,
	) (domain.Category, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Category, error)

	GetByCode(
		context.Context,
		string,
	) (domain.Category, error)

	List(
		context.Context,
	) ([]domain.Category, error)

	Update(
		context.Context,
		uuid.UUID,
		service.UpdateCategoryInput,
	) (domain.Category, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.Category, error)
}

type CategoryHandler struct {
	service categoryService
	log     zerolog.Logger
}

func NewCategoryHandler(
	service categoryService,
	log zerolog.Logger,
) *CategoryHandler {
	return &CategoryHandler{
		service: service,
		log:     log,
	}
}

type createCategoryRequest struct {
	ParentID    *uuid.UUID `json:"parent_id"`
	Name        string     `json:"name"`
	Code        string     `json:"code"`
	Description *string    `json:"description"`
}

type updateCategoryRequest struct {
	ParentID    *uuid.UUID `json:"parent_id"`
	Name        string     `json:"name"`
	Code        string     `json:"code"`
	Description *string    `json:"description"`
}

type setCategoryActiveRequest struct {
	IsActive *bool `json:"is_active"`
}

type categoryResponse struct {
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Name        string     `json:"name"`
	Code        string     `json:"code"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var request createCategoryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	category, err := h.service.Create(
		c.Request.Context(),
		service.CreateCategoryInput{
			ParentID:    request.ParentID,
			Name:        request.Name,
			Code:        request.Code,
			Description: request.Description,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create category",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toCategoryResponse(category),
	)
}

func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid category id"},
		)
		return
	}

	category, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get category",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toCategoryResponse(category),
	)
}

func (h *CategoryHandler) GetByCode(c *gin.Context) {
	category, err := h.service.GetByCode(
		c.Request.Context(),
		c.Param("code"),
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get category by code",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toCategoryResponse(category),
	)
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.service.List(
		c.Request.Context(),
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list categories",
			err,
		)
		return
	}

	response := make(
		[]categoryResponse,
		0,
		len(categories),
	)

	for _, category := range categories {
		response = append(
			response,
			toCategoryResponse(category),
		)
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid category id"},
		)
		return
	}

	var request updateCategoryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	category, err := h.service.Update(
		c.Request.Context(),
		id,
		service.UpdateCategoryInput{
			ParentID:    request.ParentID,
			Name:        request.Name,
			Code:        request.Code,
			Description: request.Description,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update category",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toCategoryResponse(category),
	)
}

func (h *CategoryHandler) SetActive(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid category id"},
		)
		return
	}

	var request setCategoryActiveRequest

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

	category, err := h.service.SetActive(
		c.Request.Context(),
		id,
		*request.IsActive,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"set category active",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toCategoryResponse(category),
	)
}

func toCategoryResponse(
	category domain.Category,
) categoryResponse {
	return categoryResponse{
		ID:          category.ID,
		ParentID:    category.ParentID,
		Name:        category.Name,
		Code:        category.Code,
		Description: category.Description,
		IsActive:    category.IsActive,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}
