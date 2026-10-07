package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterPriceRoutes(
	r *gin.Engine,
	h *handler.PriceHandler,
) {
	prices := r.Group(
		"/api/v1/catalog/prices",
	)

	prices.POST("", h.Create)
	prices.GET("", h.List)
	prices.GET("/:id", h.GetByID)
	prices.PUT("/:id", h.Update)

	r.GET(
		"/api/v1/catalog/price-lists/:price_list_id/prices",
		h.ListByPriceListID,
	)

	r.GET(
		"/api/v1/catalog/product-variants/:variant_id/prices",
		h.ListByVariantID,
	)
}
