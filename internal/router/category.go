package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

//Defines the public Category HTTP endpoints and connects them to CategoryHandler

func RegisterCategoryRoutes(
	r *gin.Engine,
	h *handler.CategoryHandler,
) {
	categories := r.Group("/api/v1/catalog/categories")

	categories.POST("", h.Create)
	categories.GET("", h.List)
	categories.GET("/:id", h.GetByID)
	categories.PUT("/:id", h.Update)
	categories.PATCH("/:id/status", h.SetActive)
}
