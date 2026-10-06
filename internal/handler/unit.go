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

type unitService interface {
	Create(
		context.Context,
		service.CreateUnitInput,
	) (domain.Unit, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Unit, error)

	List(
		context.Context,
	) ([]domain.Unit, error)

	Update(
		context.Context,
		uuid.UUID,
		service.UpdateUnitInput,
	) (domain.Unit, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.Unit, error)
}

type UnitHandler struct {
	service unitService
	log     zerolog.Logger
}

func NewUnitHandler(
	service unitService,
	log zerolog.Logger,
) *UnitHandler {
	return &UnitHandler{
		service: service,
		log:     log,
	}
}

type createUnitRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type updateUnitRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type setUnitActiveRequest struct {
	IsActive *bool `json:"is_active"`
}

type unitResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *UnitHandler) Create(c *gin.Context) {
	var request createUnitRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	unit, err := h.service.Create(
		c.Request.Context(),
		service.CreateUnitInput{
			Code: request.Code,
			Name: request.Name,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"create unit",
			err,
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		toUnitResponse(unit),
	)
}

func (h *UnitHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid unit id"},
		)
		return
	}

	unit, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get unit",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toUnitResponse(unit),
	)
}

func (h *UnitHandler) List(c *gin.Context) {
	units, err := h.service.List(
		c.Request.Context(),
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"list units",
			err,
		)
		return
	}

	response := make(
		[]unitResponse,
		0,
		len(units),
	)

	for _, unit := range units {
		response = append(
			response,
			toUnitResponse(unit),
		)
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

func (h *UnitHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid unit id"},
		)
		return
	}

	var request updateUnitRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	unit, err := h.service.Update(
		c.Request.Context(),
		id,
		service.UpdateUnitInput{
			Code: request.Code,
			Name: request.Name,
		},
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"update unit",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toUnitResponse(unit),
	)
}

func (h *UnitHandler) SetActive(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid unit id"},
		)
		return
	}

	var request setUnitActiveRequest

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

	unit, err := h.service.SetActive(
		c.Request.Context(),
		id,
		*request.IsActive,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"set unit active",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		toUnitResponse(unit),
	)
}

func toUnitResponse(
	unit domain.Unit,
) unitResponse {
	return unitResponse{
		ID:        unit.ID,
		Code:      unit.Code,
		Name:      unit.Name,
		IsActive:  unit.IsActive,
		CreatedAt: unit.CreatedAt,
		UpdatedAt: unit.UpdatedAt,
	}
}
