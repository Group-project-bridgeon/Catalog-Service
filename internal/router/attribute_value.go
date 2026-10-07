package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterAttributeValueRoutes(
	r *gin.Engine,
	h *handler.AttributeValueHandler,
) {
	values := r.Group(
		"/api/v1/catalog/attribute-values",
	)

	values.POST("", h.Create)
	values.GET("", h.List)
	values.GET("/:id", h.GetByID)
	values.PUT("/:id", h.Update)

	r.GET(
		"/api/v1/catalog/attributes/:attribute_id/values",
		h.ListByAttributeID,
	)
}
