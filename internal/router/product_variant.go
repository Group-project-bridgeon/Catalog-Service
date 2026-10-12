package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProductVariantRoutes(
	r *gin.Engine,
	h *handler.ProductVariantHandler,
) {
	variants := r.Group(
		"/api/v1/catalog/product-variants",
	)

	variants.POST("", h.Create)
	variants.GET("", h.List)
	variants.GET("/:id", h.GetByID)
	variants.PUT("/:id", h.Update)
	variants.PATCH("/:id/status", h.SetActive)

	r.GET(
		"/api/v1/catalog/products/:id/variants",
		h.ListByProductID,
	)
}
