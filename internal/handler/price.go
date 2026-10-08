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

type priceService interface {
	Create(
		context.Context,
		service.CreatePriceInput,
	) (domain.Price, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Price, error)

	List(
		context.Context,
		service.ListPricesInput,
	) (service.Page[domain.Price], error)

	ListByPriceListID(
		context.Context,
		uuid.UUID,
		service.PageInput,
	) (service.Page[domain.Price], error)

	ListByVariantID(
		context.Context,
		uuid.UUID,
		service.PageInput,
	) (service.Page[domain.Price], error)

	Update(
		context.Context,
		uuid.UUID,
		service.UpdatePriceInput,
	) (domain.Price, error)
}

type PriceHandler struct {
	service priceService
	log     zerolog.Logger
}

func NewPriceHandler(
	service priceService,
	log zerolog.Logger,
) *PriceHandler {
	return &PriceHandler{
		service: service,
		log:     log,
	}
}

type createPriceRequest struct {
	PriceListID uuid.UUID `json:"price_list_id"`
	VariantID   uuid.UUID `json:"variant_id"`
	UnitPrice   string    `json:"unit_price"`
}

type updatePriceRequest struct {
	UnitPrice string `json:"unit_price"`
}

type priceResponse struct {
	ID          uuid.UUID `json:"id"`
	PriceListID uuid.UUID `json:"price_list_id"`
	VariantID   uuid.UUID `json:"variant_id"`
	UnitPrice   string    `json:"unit_price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type pricePageResponse struct {
	Items      []priceResponse `json:"items"`
	NextCursor *string         `json:"next_cursor"`
}

func (h *PriceHandler) Create(
	c *gin.Context,
) {
	var request createPriceRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	price, err := h.service.Create(
		c.Request.Context(),
		service.CreatePriceInput{
			PriceListID: request.PriceListID,
			VariantID:   request.VariantID,
			UnitPrice:   request.UnitPrice,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create price",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toPriceResponse(price),
	)
}

func (h *PriceHandler) GetByID(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid price id"},
		)
		return
	}

	price, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get price",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toPriceResponse(price),
	)
}

func (h *PriceHandler) List(
	c *gin.Context,
) {
	pageInput, err := parsePageInput(c)
	if err != nil {
		writeInvalidQuery(c, err)
		return
	}

	priceListID, err := parseOptionalUUIDQuery(
		c,
		"price_list_id",
	)
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
		service.ListPricesInput{
			Page:        pageInput,
			PriceListID: priceListID,
			VariantID:   variantID,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list prices",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		pricePageResponse{
			Items:      toPriceResponses(page.Items),
			NextCursor: page.NextCursor,
		},
	)
}

func (h *PriceHandler) ListByPriceListID(
	c *gin.Context,
) {
	priceListID, err := uuid.Parse(
		c.Param("id"),
	)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid price list id"},
		)
		return
	}

	pageInput, err := parsePageInput(c)
	if err != nil {
		writeInvalidQuery(c, err)
		return
	}

	page, err := h.service.ListByPriceListID(
		c.Request.Context(),
		priceListID,
		pageInput,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list prices by price list",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		pricePageResponse{
			Items:      toPriceResponses(page.Items),
			NextCursor: page.NextCursor,
		},
	)
}

func (h *PriceHandler) ListByVariantID(
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
			"list prices by variant",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		pricePageResponse{
			Items:      toPriceResponses(page.Items),
			NextCursor: page.NextCursor,
		},
	)
}

func (h *PriceHandler) Update(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid price id"},
		)
		return
	}

	var request updatePriceRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	price, err := h.service.Update(
		c.Request.Context(),
		id,
		service.UpdatePriceInput{
			UnitPrice: request.UnitPrice,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update price",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toPriceResponse(price),
	)
}

func toPriceResponse(
	price domain.Price,
) priceResponse {
	return priceResponse{
		ID:          price.ID,
		PriceListID: price.PriceListID,
		VariantID:   price.VariantID,
		UnitPrice:   price.UnitPrice,
		CreatedAt:   price.CreatedAt,
		UpdatedAt:   price.UpdatedAt,
	}
}

func toPriceResponses(
	prices []domain.Price,
) []priceResponse {
	response := make(
		[]priceResponse,
		0,
		len(prices),
	)

	for _, price := range prices {
		response = append(
			response,
			toPriceResponse(price),
		)
	}

	return response
}
