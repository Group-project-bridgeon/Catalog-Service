package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func Encode[T any](value T) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf(
			"marshal cursor: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func Decode[T any](
	raw string,
	destination *T,
) error {
	if raw == "" {
		return nil
	}

	if destination == nil {
		return fmt.Errorf("cursor destination is nil")
	}

	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return fmt.Errorf(
			"decode cursor: %w",
			err,
		)
	}

	if err := json.Unmarshal(data, destination); err != nil {
		return fmt.Errorf(
			"unmarshal cursor: %w",
			err,
		)
	}

	return nil
}
