package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterAttributeRoutes(
	r *gin.Engine,
	h *handler.AttributeHandler,
) {
	attributes := r.Group(
		"/api/v1/catalog/attributes",
	)

	attributes.POST("", h.Create)
	attributes.GET("", h.List)
	attributes.GET("/:id", h.GetByID)
	attributes.PUT("/:id", h.Update)
}
