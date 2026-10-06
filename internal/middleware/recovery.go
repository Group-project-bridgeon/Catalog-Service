package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func Recovery(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			log.Error().
				Str(
					"panic_type",
					fmt.Sprintf("%T", recovered),
				).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("stack", string(debug.Stack())).
				Msg("panic recovered")

			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "internal server error",
				},
			)
		}()

		c.Next()
	}
}
