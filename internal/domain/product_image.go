package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProductImage struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	VariantID *uuid.UUID
	ImageURL  string
	SortOrder int32
	CreatedAt time.Time
}
