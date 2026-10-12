package router

import (
	"testing"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/gin-gonic/gin"
)

func TestAllRoutesRegisterWithoutPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	RegisterCategoryRoutes(r, (*handler.CategoryHandler)(nil))
	RegisterBrandRoutes(r, (*handler.BrandHandler)(nil))
	RegisterUnitRoutes(r, (*handler.UnitHandler)(nil))
	RegisterProductRoutes(r, (*handler.ProductHandler)(nil))
	RegisterProductVariantRoutes(r, (*handler.ProductVariantHandler)(nil))
	RegisterVariantValueRoutes(r, (*handler.VariantValueHandler)(nil))
	RegisterAttributeRoutes(r, (*handler.AttributeHandler)(nil))
	RegisterAttributeValueRoutes(r, (*handler.AttributeValueHandler)(nil))
	RegisterProductImageRoutes(r, (*handler.ProductImageHandler)(nil))
	RegisterPriceListRoutes(r, (*handler.PriceListHandler)(nil))
	RegisterPriceRoutes(r, (*handler.PriceHandler)(nil))
	RegisterProductDetailRoutes(r, (*handler.ProductDetailHandler)(nil))
}
