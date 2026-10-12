package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariant struct {
	ID                  uuid.UUID
	ProductID           uuid.UUID
	SKU                 string
	DefaultSellingPrice *string
	UnitID              uuid.UUID
	IsActive            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
