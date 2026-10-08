package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestWriteReferenceJSON_ETagAnd304(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zerolog.Nop()

	data := []map[string]string{
		{"id": "1", "name": "Category 1"},
	}

	r := gin.New()
	r.GET("/categories", func(c *gin.Context) {
		writeReferenceJSON(c, log, "list test", data)
	})

	// 1. Initial Request returns 200 with ETag & Cache-Control
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/categories", nil)
	r.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	etag := w1.Header().Get("ETag")
	assert.NotEmpty(t, etag)
	assert.Equal(t, referenceCacheControl, w1.Header().Get("Cache-Control"))
	assert.Contains(t, w1.Body.String(), "Category 1")

	// 2. Request with matching If-None-Match returns 304 Not Modified and empty body
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/categories", nil)
	req2.Header.Set("If-None-Match", etag)
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusNotModified, w2.Code)
	assert.Empty(t, w2.Body.String())
}
