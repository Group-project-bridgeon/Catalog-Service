package router

import (
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProductDetailRoutes(
	r *gin.Engine,
	h *handler.ProductDetailHandler,
) {
	r.GET(
		"/api/v1/catalog/products/:id/detail",
		h.Get,
	)
}
