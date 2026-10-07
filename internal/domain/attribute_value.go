package domain

import "github.com/google/uuid"

type AttributeValue struct {
	ID          uuid.UUID
	AttributeID uuid.UUID
	Value       string
}
