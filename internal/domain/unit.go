package domain

import (
	"time"

	"github.com/google/uuid"
)

type Unit struct {
	ID        uuid.UUID
	Code      string
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
