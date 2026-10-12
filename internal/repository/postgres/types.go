package postgres

import (
	"fmt"
	"math/big"
	"strings"
	"time"

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

func decimalStringToPGNumeric(
	value *string,
) (pgtype.Numeric, error) {
	if value == nil {
		return pgtype.Numeric{}, nil
	}

	raw := strings.TrimSpace(*value)

	if raw == "" {
		return pgtype.Numeric{}, fmt.Errorf(
			"decimal value is empty",
		)
	}

	negative := false

	switch raw[0] {
	case '-':
		negative = true
		raw = raw[1:]
	case '+':
		raw = raw[1:]
	}

	if raw == "" {
		return pgtype.Numeric{}, fmt.Errorf(
			"invalid decimal value",
		)
	}

	parts := strings.Split(raw, ".")

	if len(parts) > 2 ||
		parts[0] == "" ||
		(len(parts) == 2 && parts[1] == "") {
		return pgtype.Numeric{}, fmt.Errorf(
			"invalid decimal value",
		)
	}

	scale := 0
	digits := parts[0]

	if len(parts) == 2 {
		scale = len(parts[1])
		digits += parts[1]
	}

	if negative {
		digits = "-" + digits
	}

	number, ok := new(big.Int).SetString(
		digits,
		10,
	)
	if !ok {
		return pgtype.Numeric{}, fmt.Errorf(
			"invalid decimal value",
		)
	}

	return pgtype.Numeric{
		Int:   number,
		Exp:   -int32(scale),
		Valid: true,
	}, nil
}

func pgNumericToDecimalString(
	value pgtype.Numeric,
) (*string, error) {
	if !value.Valid {
		return nil, nil
	}

	if value.NaN || value.InfinityModifier != 0 {
		return nil, fmt.Errorf(
			"unsupported numeric value",
		)
	}

	if value.Int == nil {
		return nil, fmt.Errorf(
			"numeric value has no integer component",
		)
	}

	negative := value.Int.Sign() < 0

	absolute := new(big.Int).Abs(
		new(big.Int).Set(value.Int),
	)

	digits := absolute.String()
	exponent := int(value.Exp)

	var result string

	if exponent >= 0 {
		result = digits +
			strings.Repeat("0", exponent)
	} else {
		scale := -exponent

		if len(digits) <= scale {
			digits = strings.Repeat(
				"0",
				scale-len(digits)+1,
			) + digits
		}

		point := len(digits) - scale

		result = digits[:point] +
			"." +
			digits[point:]
	}

	if negative {
		result = "-" + result
	}

	return &result, nil
}

func int32ToPGInt4(
	value int32,
) pgtype.Int4 {
	return pgtype.Int4{
		Int32: value,
		Valid: true,
	}
}

func pgInt4ToInt32(
	value pgtype.Int4,
) int32 {
	if !value.Valid {
		return 0
	}

	return value.Int32
}

func boolPtrToPGBool(
	value *bool,
) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{}
	}

	return pgtype.Bool{
		Bool:  *value,
		Valid: true,
	}
}

func int32PtrToPGInt4(
	value *int32,
) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}

	return pgtype.Int4{
		Int32: *value,
		Valid: true,
	}
}

func timePtrToPGTimestamptz(
	value *time.Time,
) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{
		Time:  *value,
		Valid: true,
	}
}

func uuidSliceToPG(
	values []uuid.UUID,
) []pgtype.UUID {
	result := make(
		[]pgtype.UUID,
		0,
		len(values),
	)

	for _, value := range values {
		result = append(
			result,
			uuidToPG(value),
		)
	}

	return result
}
