package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterVariantValueRoutes(
	r *gin.Engine,
	h *handler.VariantValueHandler,
) {
	values := r.Group(
		"/api/v1/catalog/variant-values",
	)

	values.POST("", h.Create)
	values.GET("", h.List)

	values.GET(
		"/:variant_id/:attribute_id",
		h.Get,
	)

	values.PUT(
		"/:variant_id/:attribute_id",
		h.Update,
	)

	r.GET(
		"/api/v1/catalog/product-variants/:id/values",
		h.ListByVariantID,
	)
}
