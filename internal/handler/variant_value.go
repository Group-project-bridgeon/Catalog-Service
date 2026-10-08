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

type variantValueService interface {
	Create(
		context.Context,
		service.CreateVariantValueInput,
	) (domain.VariantValue, error)

	GetByVariantAndAttribute(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (domain.VariantValue, error)

	List(
		context.Context,
		service.ListVariantValuesInput,
	) (service.Page[domain.VariantValue], error)

	ListByVariantID(
		context.Context,
		uuid.UUID,
		service.PageInput,
	) (service.Page[domain.VariantValue], error)

	Update(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		service.UpdateVariantValueInput,
	) (domain.VariantValue, error)
}

type VariantValueHandler struct {
	service variantValueService
	log     zerolog.Logger
}

func NewVariantValueHandler(
	service variantValueService,
	log zerolog.Logger,
) *VariantValueHandler {
	return &VariantValueHandler{
		service: service,
		log:     log,
	}
}

type createVariantValueRequest struct {
	VariantID        uuid.UUID `json:"variant_id"`
	AttributeID      uuid.UUID `json:"attribute_id"`
	AttributeValueID uuid.UUID `json:"attribute_value_id"`
}

type updateVariantValueRequest struct {
	AttributeValueID uuid.UUID `json:"attribute_value_id"`
}

type variantValueResponse struct {
	VariantID        uuid.UUID `json:"variant_id"`
	AttributeID      uuid.UUID `json:"attribute_id"`
	AttributeValueID uuid.UUID `json:"attribute_value_id"`
}

type variantValuePageResponse struct {
	Items      []variantValueResponse `json:"items"`
	NextCursor *string                `json:"next_cursor"`
}

func (h *VariantValueHandler) Create(
	c *gin.Context,
) {
	var request createVariantValueRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	result, err := h.service.Create(
		c.Request.Context(),
		service.CreateVariantValueInput{
			VariantID:        request.VariantID,
			AttributeID:      request.AttributeID,
			AttributeValueID: request.AttributeValueID,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create variant value",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toVariantValueResponse(result),
	)
}

func (h *VariantValueHandler) Get(
	c *gin.Context,
) {
	variantID, attributeID, ok :=
		parseVariantValueParams(c)

	if !ok {
		return
	}

	result, err :=
		h.service.GetByVariantAndAttribute(
			c.Request.Context(),
			variantID,
			attributeID,
		)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get variant value",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toVariantValueResponse(result),
	)
}

func (h *VariantValueHandler) List(
	c *gin.Context,
) {
	pageInput, err := parsePageInput(c)
	if err != nil {
		writeInvalidQuery(c, err)
		return
	}

	variantID, err := parseOptionalUUIDQuery(
		c,
		"variant_id",
	)
	if err != nil {
		writeInvalidQuery(c, err)
		return
	}

	page, err := h.service.List(
		c.Request.Context(),
		service.ListVariantValuesInput{
			Page:      pageInput,
			VariantID: variantID,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list variant values",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		variantValuePageResponse{
			Items:      toVariantValueResponses(page.Items),
			NextCursor: page.NextCursor,
		},
	)
}

func (h *VariantValueHandler) ListByVariantID(
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

	pageInput, err := parsePageInput(c)
	if err != nil {
		writeInvalidQuery(c, err)
		return
	}

	page, err := h.service.ListByVariantID(
		c.Request.Context(),
		variantID,
		pageInput,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list variant values by variant",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		variantValuePageResponse{
			Items:      toVariantValueResponses(page.Items),
			NextCursor: page.NextCursor,
		},
	)
}

func (h *VariantValueHandler) Update(
	c *gin.Context,
) {
	variantID, attributeID, ok :=
		parseVariantValueParams(c)

	if !ok {
		return
	}

	var request updateVariantValueRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	result, err := h.service.Update(
		c.Request.Context(),
		variantID,
		attributeID,
		service.UpdateVariantValueInput{
			AttributeValueID: request.AttributeValueID,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update variant value",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toVariantValueResponse(result),
	)
}

func parseVariantValueParams(
	c *gin.Context,
) (uuid.UUID, uuid.UUID, bool) {
	variantID, err := uuid.Parse(
		c.Param("variant_id"),
	)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid variant id"},
		)
		return uuid.Nil, uuid.Nil, false
	}

	attributeID, err := uuid.Parse(
		c.Param("attribute_id"),
	)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid attribute id"},
		)
		return uuid.Nil, uuid.Nil, false
	}

	return variantID, attributeID, true
}

func toVariantValueResponse(
	value domain.VariantValue,
) variantValueResponse {
	return variantValueResponse{
		VariantID:        value.VariantID,
		AttributeID:      value.AttributeID,
		AttributeValueID: value.AttributeValueID,
	}
}

func toVariantValueResponses(
	values []domain.VariantValue,
) []variantValueResponse {
	response := make(
		[]variantValueResponse,
		0,
		len(values),
	)

	for _, value := range values {
		response = append(
			response,
			toVariantValueResponse(value),
		)
	}

	return response
}
