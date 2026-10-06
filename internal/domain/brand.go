package domain

import (
	"time"

	"github.com/google/uuid"
)

type Brand struct {
	ID        uuid.UUID
	Name      string
	Code      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
