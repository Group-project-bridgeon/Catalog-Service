package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const referenceCacheControl = "public, max-age=30, stale-while-revalidate=30"

func writeReferenceJSON(
	c *gin.Context,
	log zerolog.Logger,
	operation string,
	value any,
) {
	body, err := json.Marshal(value)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", operation).
			Msg("marshal reference response")

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	sum := sha256.Sum256(body)

	etag := `"` +
		hex.EncodeToString(sum[:]) +
		`"`

	c.Header(
		"Cache-Control",
		referenceCacheControl,
	)

	c.Header(
		"ETag",
		etag,
	)

	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.Data(
		http.StatusOK,
		"application/json; charset=utf-8",
		body,
	)
}
