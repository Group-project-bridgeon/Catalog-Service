package router

import (
	"context"
	"net/http"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type DatabaseHealthChecker interface {
	Ping(context.Context) error
}

func New(
	db DatabaseHealthChecker,
	log zerolog.Logger,
) *gin.Engine {
	r := gin.New()

	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.Recovery(log))

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

	return r
}
