package domain

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	BrandID     *uuid.UUID
	Name        string
	ProductCode string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
