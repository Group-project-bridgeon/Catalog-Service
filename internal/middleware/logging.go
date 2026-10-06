package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func RequestLogger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()

		event := log.Info()

		switch {
		case status >= 500:
			event = log.Error()

		case status >= 400:
			event = log.Warn()
		}

		event.
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", status).
			Int64(
				"duration_ms",
				time.Since(start).Milliseconds(),
			)

		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			event.Str("request_id", requestID)
		}

		if correlationID := c.GetHeader("X-Correlation-ID"); correlationID != "" {
			event.Str("correlation_id", correlationID)
		}

		event.Msg("http request")
	}
}
