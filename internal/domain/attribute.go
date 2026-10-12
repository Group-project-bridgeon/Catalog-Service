package domain

import "github.com/google/uuid"

type Attribute struct {
	ID   uuid.UUID
	Name string
	Code string
}
