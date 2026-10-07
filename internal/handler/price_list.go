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

type priceListService interface {
	Create(
		context.Context,
		service.CreatePriceListInput,
	) (domain.PriceList, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.PriceList, error)

	List(
		context.Context,
	) ([]domain.PriceList, error)

	Update(
		context.Context,
		uuid.UUID,
		service.UpdatePriceListInput,
	) (domain.PriceList, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.PriceList, error)
}

type PriceListHandler struct {
	service priceListService
	log     zerolog.Logger
}

func NewPriceListHandler(
	service priceListService,
	log zerolog.Logger,
) *PriceListHandler {
	return &PriceListHandler{
		service: service,
		log:     log,
	}
}

type createPriceListRequest struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	CurrencyCode string `json:"currency_code"`
}

type updatePriceListRequest struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	CurrencyCode string `json:"currency_code"`
}

type setPriceListActiveRequest struct {
	IsActive *bool `json:"is_active"`
}

type priceListResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	CurrencyCode string    `json:"currency_code"`
	IsActive     bool      `json:"is_active"`
}

func (h *PriceListHandler) Create(
	c *gin.Context,
) {
	var request createPriceListRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	priceList, err := h.service.Create(
		c.Request.Context(),
		service.CreatePriceListInput{
			Name:         request.Name,
			Code:         request.Code,
			CurrencyCode: request.CurrencyCode,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create price list",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toPriceListResponse(priceList),
	)
}

func (h *PriceListHandler) GetByID(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid price list id"},
		)
		return
	}

	priceList, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get price list",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toPriceListResponse(priceList),
	)
}

func (h *PriceListHandler) List(
	c *gin.Context,
) {
	priceLists, err := h.service.List(
		c.Request.Context(),
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list price lists",
			err,
		)
		return
	}

	response := make(
		[]priceListResponse,
		0,
		len(priceLists),
	)

	for _, priceList := range priceLists {
		response = append(
			response,
			toPriceListResponse(priceList),
		)
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

func (h *PriceListHandler) Update(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid price list id"},
		)
		return
	}

	var request updatePriceListRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	priceList, err := h.service.Update(
		c.Request.Context(),
		id,
		service.UpdatePriceListInput{
			Name:         request.Name,
			Code:         request.Code,
			CurrencyCode: request.CurrencyCode,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update price list",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toPriceListResponse(priceList),
	)
}

func (h *PriceListHandler) SetActive(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid price list id"},
		)
		return
	}

	var request setPriceListActiveRequest

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

	priceList, err := h.service.SetActive(
		c.Request.Context(),
		id,
		*request.IsActive,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"set price list active",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toPriceListResponse(priceList),
	)
}

func toPriceListResponse(
	priceList domain.PriceList,
) priceListResponse {
	return priceListResponse{
		ID:           priceList.ID,
		Name:         priceList.Name,
		Code:         priceList.Code,
		CurrencyCode: priceList.CurrencyCode,
		IsActive:     priceList.IsActive,
	}
}
