package handler

import (
	"errors"
	"net/http"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func writeServiceError(
	c *gin.Context,
	log zerolog.Logger,
	operation string,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrInvalidArgument):
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request"},
		)

	case errors.Is(err, service.ErrNotFound):
		c.JSON(
			http.StatusNotFound,
			gin.H{"error": "resource not found"},
		)

	case errors.Is(err, service.ErrConflict):
		c.JSON(
			http.StatusConflict,
			gin.H{"error": "resource already exists"},
		)

	case errors.Is(err, service.ErrInvalidReference):
		c.JSON(
			http.StatusUnprocessableEntity,
			gin.H{"error": "invalid reference"},
		)

	default:
		event := log.Error().
			Err(err).
			Str("operation", operation)

		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			event.Str("request_id", requestID)
		}

		if correlationID := c.GetHeader("X-Correlation-ID"); correlationID != "" {
			event.Str("correlation_id", correlationID)
		}

		event.Msg("request failed")

		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
		)
	}
}
