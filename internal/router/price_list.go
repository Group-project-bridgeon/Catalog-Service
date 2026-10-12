package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterPriceListRoutes(
	r *gin.Engine,
	h *handler.PriceListHandler,
) {
	priceLists := r.Group(
		"/api/v1/catalog/price-lists",
	)

	priceLists.POST("", h.Create)
	priceLists.GET("", h.List)
	priceLists.GET("/:id", h.GetByID)
	priceLists.PUT("/:id", h.Update)
	priceLists.PATCH("/:id/status", h.SetActive)
}
