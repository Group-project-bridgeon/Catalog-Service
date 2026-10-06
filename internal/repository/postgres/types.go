package postgres

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func uuidToPG(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: [16]byte(id),
		Valid: true,
	}
}

func nullableUUIDToPG(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}

	return uuidToPG(*id)
}

func pgUUIDToUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}

func pgUUIDToNullableUUID(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}

	value := uuid.UUID(id.Bytes)

	return &value
}

func stringPtrToPGText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}

func pgTextToStringPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	result := value.String

	return &result
}
