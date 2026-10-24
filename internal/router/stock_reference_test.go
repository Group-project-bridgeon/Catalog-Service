package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"strings"
	"testing"
	"time"
)

func TestStockReferenceTransportDisabledByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New(nil, zerolog.Nop(), 1024, time.Second, nil)
	registerBusinessRoutes(r)
	for _, route := range r.Routes() {
		if strings.Contains(route.Path, "stock-reference") || strings.Contains(route.Path, "pack") || strings.Contains(route.Path, "eligibility") {
			t.Fatal("unapproved route registered")
		}
	}
	for key := range PermissionMatrix() {
		if strings.Contains(key, "stock-reference") {
			t.Fatal("permission policy broadened")
		}
	}
}
