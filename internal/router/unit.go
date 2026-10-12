package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterUnitRoutes(
	r *gin.Engine,
	h *handler.UnitHandler,
) {
	units := r.Group("/api/v1/catalog/units")

	units.POST("", h.Create)
	units.GET("", h.List)
	units.GET("/:id", h.GetByID)
	units.PUT("/:id", h.Update)
	units.PATCH("/:id/status", h.SetActive)
}
