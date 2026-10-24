package router

import (
	"context"
	"net/http"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/middleware"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type DatabaseHealthChecker interface {
	Ping(context.Context) error
}

const StockReferencePath = "/api/v1/catalog/stock-references/batch"

type StockReferenceTransport struct {
	Verifier *security.MachineVerifier
	Resolver handler.StockReferenceResolver
}

func New(
	db DatabaseHealthChecker,
	log zerolog.Logger,
	maxBodyBytes int64,
	slowRequestThreshold time.Duration,
	verifier *security.Verifier,
	stock ...StockReferenceTransport,
) *gin.Engine {
	r := gin.New()

	r.Use(
		middleware.RequestLogger(
			log,
			slowRequestThreshold,
		),
	)
	r.Use(middleware.Recovery(log))
	human := security.Protect(verifier, PermissionMatrix())
	r.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodPost && c.FullPath() == StockReferencePath && len(stock) == 1 && stock[0].Verifier != nil && stock[0].Resolver != nil {
			c.Header("Cache-Control", "no-store")
			security.ProtectMachine(stock[0].Verifier)(c)
			if c.Writer.Status() == 401 || c.Writer.Status() == 403 {
				log.Warn().Str("component", "catalog_reference").Str("operation", "authenticate").Str("reason", "machine_access_denied").Str("request_id", uuid.NewString()).Msg("reference request rejected")
			}
			return
		}
		human(c)
	})
	r.Use(middleware.MaxBodyBytes(maxBodyBytes))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"status":  "ok",
				"service": "catalog",
			},
		)
	})

	r.GET("/ready", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(
				http.StatusServiceUnavailable,
				gin.H{
					"status":  "not_ready",
					"service": "catalog",
				},
			)

			return
		}

		c.JSON(
			http.StatusOK,
			gin.H{
				"status":  "ready",
				"service": "catalog",
			},
		)
	})

	if len(stock) == 1 && stock[0].Verifier != nil && stock[0].Resolver != nil {
		r.POST(StockReferencePath, handler.StockReferenceBatch(stock[0].Resolver, log))
	}
	return r
}
