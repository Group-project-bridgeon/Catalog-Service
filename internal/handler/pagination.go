package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func parsePageInput(
	c *gin.Context,
) (service.PageInput, error) {
	rawLimit := strings.TrimSpace(
		c.Query("limit"),
	)

	limit := service.DefaultPageLimit

	if rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			return service.PageInput{}, fmt.Errorf(
				"invalid limit",
			)
		}

		limit = parsed
	}

	if limit < 1 ||
		limit > service.MaxPageLimit {
		return service.PageInput{}, fmt.Errorf(
			"limit must be between 1 and %d",
			service.MaxPageLimit,
		)
	}

	return service.PageInput{
		Limit:  limit,
		Cursor: strings.TrimSpace(c.Query("cursor")),
	}, nil
}

func parseOptionalUUIDQuery(
	c *gin.Context,
	name string,
) (*uuid.UUID, error) {
	raw := strings.TrimSpace(c.Query(name))

	if raw == "" {
		return nil, nil
	}

	value, err := uuid.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid %s",
			name,
		)
	}

	return &value, nil
}

func parseOptionalBoolQuery(
	c *gin.Context,
	name string,
) (*bool, error) {
	raw := strings.TrimSpace(c.Query(name))

	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid %s",
			name,
		)
	}

	return &value, nil
}

func parseUUIDList(
	raw string,
	max int,
) ([]uuid.UUID, error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return nil, fmt.Errorf(
			"ids cannot be empty",
		)
	}

	parts := strings.Split(raw, ",")

	if len(parts) > max {
		return nil, fmt.Errorf(
			"maximum %d ids are allowed",
			max,
		)
	}

	result := make(
		[]uuid.UUID,
		0,
		len(parts),
	)

	seen := make(map[uuid.UUID]struct{})

	for _, part := range parts {
		id, err := uuid.Parse(
			strings.TrimSpace(part),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid variant id",
			)
		}

		if _, exists := seen[id]; exists {
			continue
		}

		seen[id] = struct{}{}
		result = append(result, id)
	}

	return result, nil
}

func writeInvalidQuery(
	c *gin.Context,
	err error,
) {
	c.JSON(
		400,
		gin.H{
			"error": err.Error(),
		},
	)
}
