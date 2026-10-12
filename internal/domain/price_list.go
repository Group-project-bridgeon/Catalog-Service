package domain

import "github.com/google/uuid"

type PriceList struct {
	ID           uuid.UUID
	Name         string
	Code         string
	CurrencyCode string
	IsActive     bool
}
