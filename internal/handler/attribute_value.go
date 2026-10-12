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

type attributeValueService interface {
	Create(
		context.Context,
		service.CreateAttributeValueInput,
	) (domain.AttributeValue, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.AttributeValue, error)

	List(
		context.Context,
	) ([]domain.AttributeValue, error)

	ListByAttributeID(
		context.Context,
		uuid.UUID,
	) ([]domain.AttributeValue, error)

	Update(
		context.Context,
		uuid.UUID,
		service.UpdateAttributeValueInput,
	) (domain.AttributeValue, error)
}

type AttributeValueHandler struct {
	service attributeValueService
	log     zerolog.Logger
}

func NewAttributeValueHandler(
	service attributeValueService,
	log zerolog.Logger,
) *AttributeValueHandler {
	return &AttributeValueHandler{
		service: service,
		log:     log,
	}
}

type createAttributeValueRequest struct {
	AttributeID uuid.UUID `json:"attribute_id"`
	Value       string    `json:"value"`
}

type updateAttributeValueRequest struct {
	Value string `json:"value"`
}

type attributeValueResponse struct {
	ID          uuid.UUID `json:"id"`
	AttributeID uuid.UUID `json:"attribute_id"`
	Value       string    `json:"value"`
}

func (h *AttributeValueHandler) Create(
	c *gin.Context,
) {
	var request createAttributeValueRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	value, err := h.service.Create(
		c.Request.Context(),
		service.CreateAttributeValueInput{
			AttributeID: request.AttributeID,
			Value:       request.Value,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create attribute value",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toAttributeValueResponse(value),
	)
}

func (h *AttributeValueHandler) GetByID(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid attribute value id"},
		)
		return
	}

	value, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get attribute value",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toAttributeValueResponse(value),
	)
}

func (h *AttributeValueHandler) List(
	c *gin.Context,
) {
	values, err := h.service.List(
		c.Request.Context(),
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list attribute values",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toAttributeValueResponses(values),
	)
}

func (h *AttributeValueHandler) ListByAttributeID(
	c *gin.Context,
) {
	attributeID, err := uuid.Parse(
		c.Param("id"),
	)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid attribute id"},
		)
		return
	}

	values, err := h.service.ListByAttributeID(
		c.Request.Context(),
		attributeID,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list attribute values by attribute",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toAttributeValueResponses(values),
	)
}

func (h *AttributeValueHandler) Update(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid attribute value id"},
		)
		return
	}

	var request updateAttributeValueRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	value, err := h.service.Update(
		c.Request.Context(),
		id,
		service.UpdateAttributeValueInput{
			Value: request.Value,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update attribute value",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toAttributeValueResponse(value),
	)
}

func toAttributeValueResponse(
	value domain.AttributeValue,
) attributeValueResponse {
	return attributeValueResponse{
		ID:          value.ID,
		AttributeID: value.AttributeID,
		Value:       value.Value,
	}
}

func toAttributeValueResponses(
	values []domain.AttributeValue,
) []attributeValueResponse {
	response := make(
		[]attributeValueResponse,
		0,
		len(values),
	)

	for _, value := range values {
		response = append(
			response,
			toAttributeValueResponse(value),
		)
	}

	return response
}
