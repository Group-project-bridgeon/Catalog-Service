package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProductRoutes(
	r *gin.Engine,
	h *handler.ProductHandler,
) {
	products := r.Group("/api/v1/catalog/products")

	products.POST("", h.Create)
	products.GET("", h.List)
	products.GET("/:id", h.GetByID)
	products.PUT("/:id", h.Update)
	products.PATCH("/:id/status", h.SetActive)
}
