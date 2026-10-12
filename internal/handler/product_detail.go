package handler

import (
	"context"
	"net/http"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type productDetailService interface {
	Get(
		context.Context,
		uuid.UUID,
	) (domain.ProductDetail, error)
}

type ProductDetailHandler struct {
	service productDetailService
	log     zerolog.Logger
}

func NewProductDetailHandler(
	service productDetailService,
	log zerolog.Logger,
) *ProductDetailHandler {
	return &ProductDetailHandler{
		service: service,
		log:     log,
	}
}

type productDetailResponse struct {
	Product       productResponse          `json:"product"`
	Variants      []productVariantResponse `json:"variants"`
	VariantValues []variantValueResponse   `json:"variant_values"`
	Images        []productImageResponse   `json:"images"`
	Prices        []priceResponse          `json:"prices"`
}

func (h *ProductDetailHandler) Get(
	c *gin.Context,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid product id"},
		)
		return
	}

	detail, err := h.service.Get(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(
			c,
			h.log,
			"get product detail",
			err,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		productDetailResponse{
			Product:       toProductResponse(detail.Product),
			Variants:      toProductVariantResponses(detail.Variants),
			VariantValues: toVariantValueResponses(detail.VariantValues),
			Images:        toProductImageResponses(detail.Images),
			Prices:        toPriceResponses(detail.Prices),
		},
	)
}
