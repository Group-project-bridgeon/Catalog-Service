package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProductImageRoutes(
	r *gin.Engine,
	h *handler.ProductImageHandler,
) {
	images := r.Group(
		"/api/v1/catalog/product-images",
	)

	images.POST("", h.Create)
	images.GET("", h.List)
	images.GET("/:id", h.GetByID)
	images.PUT("/:id", h.Update)

	r.GET(
		"/api/v1/catalog/products/:id/images",
		h.ListByProductID,
	)

	r.GET(
		"/api/v1/catalog/product-variants/:id/images",
		h.ListByVariantID,
	)
}
