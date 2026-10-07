package handler

import (
	"context"
	"net/http"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type attributeService interface {
	Create(
		context.Context,
		service.CreateAttributeInput,
	) (domain.Attribute, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Attribute, error)

	List(
		context.Context,
	) ([]domain.Attribute, error)

	Update(
		context.Context,
		uuid.UUID,
		service.UpdateAttributeInput,
	) (domain.Attribute, error)
}

type AttributeHandler struct {
	service attributeService
	log     zerolog.Logger
}

func NewAttributeHandler(
	service attributeService,
	log zerolog.Logger,
) *AttributeHandler {
	return &AttributeHandler{
		service: service,
		log:     log,
	}
}

type createAttributeRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type updateAttributeRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type attributeResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

func (h *AttributeHandler) Create(
	c *gin.Context,
) {
	var request createAttributeRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	attribute, err := h.service.Create(
		c.Request.Context(),
		service.CreateAttributeInput{
			Name: request.Name,
			Code: request.Code,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create attribute",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toAttributeResponse(attribute),
	)
}

func (h *AttributeHandler) GetByID(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid attribute id"},
		)
		return
	}

	attribute, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get attribute",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toAttributeResponse(attribute),
	)
}

func (h *AttributeHandler) List(
	c *gin.Context,
) {
	attributes, err := h.service.List(
		c.Request.Context(),
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list attributes",
			err,
		)
		return
	}

	response := make(
		[]attributeResponse,
		0,
		len(attributes),
	)

	for _, attribute := range attributes {
		response = append(
			response,
			toAttributeResponse(attribute),
		)
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

func (h *AttributeHandler) Update(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid attribute id"},
		)
		return
	}

	var request updateAttributeRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	attribute, err := h.service.Update(
		c.Request.Context(),
		id,
		service.UpdateAttributeInput{
			Name: request.Name,
			Code: request.Code,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update attribute",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toAttributeResponse(attribute),
	)
}

func toAttributeResponse(
	attribute domain.Attribute,
) attributeResponse {
	return attributeResponse{
		ID:   attribute.ID,
		Name: attribute.Name,
		Code: attribute.Code,
	}
}
