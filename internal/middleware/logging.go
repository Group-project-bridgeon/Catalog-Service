package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func RequestLogger(
	log zerolog.Logger,
	slowThreshold time.Duration,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		path := c.Request.URL.Path

		if path == "/health" ||
			path == "/ready" {
			return
		}

		status := c.Writer.Status()
		duration := time.Since(start)

		var event *zerolog.Event

		switch {
		case status >= 500:
			event = log.Error()

		case status >= 400:
			event = log.Warn()

		case duration >= slowThreshold:
			event = log.Info()

		default:
			return
		}

		event.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Int64(
				"duration_ms",
				duration.Milliseconds(),
			)

		if requestID := c.GetHeader(
			"X-Request-ID",
		); requestID != "" {
			event.Str(
				"request_id",
				requestID,
			)
		}

		if correlationID := c.GetHeader(
			"X-Correlation-ID",
		); correlationID != "" {
			event.Str(
				"correlation_id",
				correlationID,
			)
		}

		event.Msg("http request")
	}
}
