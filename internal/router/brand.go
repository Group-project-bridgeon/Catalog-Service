package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

//Defines the public Brand HTTP endpoints and connects them to BrandHandler

func RegisterBrandRoutes(
	r *gin.Engine,
	h *handler.BrandHandler,
) {
	brands := r.Group("/api/v1/catalog/brands")

	brands.POST("", h.Create)
	brands.GET("", h.List)
	brands.GET("/:id", h.GetByID)
	brands.PUT("/:id", h.Update)
	brands.PATCH("/:id/status", h.SetActive)
}
