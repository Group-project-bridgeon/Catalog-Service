package domain

import "github.com/google/uuid"

type VariantValue struct {
	VariantID        uuid.UUID
	AttributeID      uuid.UUID
	AttributeValueID uuid.UUID
}
