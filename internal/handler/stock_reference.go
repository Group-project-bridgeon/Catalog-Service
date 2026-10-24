package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/strictjson"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type StockReferenceResolver interface {
	ResolveBatch(context.Context, []domain.StockReferenceRequest) (domain.StockReferenceBatch, error)
}

func StockReferenceBatch(resolver StockReferenceResolver, log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		fail := func(status int, reason string, err error) {
			id := uuid.NewString()
			c.Header("X-Request-ID", id)
			event := log.Warn().Str("component", "catalog_reference").Str("operation", "resolve_batch").Str("reason", reason).Str("request_id", id)
			var db interface{ SQLStateCode() string }
			if errors.As(err, &db) && db.SQLStateCode() != "" {
				event = event.Str("sqlstate", db.SQLStateCode())
			}
			event.Msg("reference request rejected")
			c.AbortWithStatusJSON(status, gin.H{"error": "stock reference request failed"})
		}
		media, params, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
		if c.Request.TLS == nil || c.Request.URL.RawQuery != "" || len(c.Request.Header.Values("Content-Type")) != 1 || err != nil || media != "application/json" || len(params) != 0 {
			fail(400, "invalid_transport", nil)
			return
		}
		data, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10))
		if err != nil {
			fail(400, "request_limit", nil)
			return
		}
		var input struct {
			Items []struct {
				VariantID        string  `json:"variant_id"`
				PackID           *string `json:"pack_id,omitempty"`
				PackRevision     *string `json:"pack_revision,omitempty"`
				ExpectedRevision *string `json:"expected_revision,omitempty"`
			} `json:"items"`
		}
		if strictjson.Decode(data, &input) != nil {
			fail(400, "invalid_json", nil)
			return
		}
		// Null is not a meaningful request value, including optional fields.
		var raw any
		if json.Unmarshal(data, &raw) != nil {
			fail(400, "invalid_json", nil)
			return
		}
		var nulls func(any) bool
		nulls = func(v any) bool {
			switch x := v.(type) {
			case nil:
				return true
			case map[string]any:
				for _, v := range x {
					if nulls(v) {
						return true
					}
				}
			case []any:
				for _, v := range x {
					if nulls(v) {
						return true
					}
				}
			}
			return false
		}
		if nulls(raw) {
			fail(400, "null_field", nil)
			return
		}
		items := make([]domain.StockReferenceRequest, 0, len(input.Items))
		for _, item := range input.Items {
			id, err := uuid.Parse(item.VariantID)
			if err != nil || id == uuid.Nil || id.String() != item.VariantID {
				fail(400, "variant_id", nil)
				return
			}
			request := domain.StockReferenceRequest{VariantID: id}
			if item.PackID != nil {
				pack, err := uuid.Parse(*item.PackID)
				if err != nil || pack == uuid.Nil || pack.String() != *item.PackID {
					fail(400, "pack_id", nil)
					return
				}
				request.PackID = pack
			}
			if item.PackRevision != nil {
				request.PackRevision = *item.PackRevision
			}
			if item.ExpectedRevision != nil {
				request.ExpectedRevision = *item.ExpectedRevision
			}
			items = append(items, request)
		}
		if domain.ValidateStockRequests(items) != nil {
			fail(400, "invalid_batch", nil)
			return
		}
		batch, err := resolver.ResolveBatch(c.Request.Context(), items)
		if err != nil {
			fail(503, "reference_unavailable", err)
			return
		}
		c.JSON(200, batch)
	}
}
