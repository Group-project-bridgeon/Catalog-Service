package domain

import (
	"time"

	"github.com/google/uuid"
)

type Price struct {
	ID          uuid.UUID
	PriceListID uuid.UUID
	VariantID   uuid.UUID
	UnitPrice   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
