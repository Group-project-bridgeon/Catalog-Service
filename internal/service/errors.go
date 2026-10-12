package service

import (
	"errors"
	"fmt"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
)

var (
	ErrInvalidArgument  = errors.New("service: invalid argument")
	ErrNotFound         = errors.New("service: not found")
	ErrConflict         = errors.New("service: conflict")
	ErrInvalidReference = errors.New("service: invalid reference")
)

func mapRepositoryError(operation string, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return fmt.Errorf("%s: %w", operation, ErrNotFound)

	case errors.Is(err, repository.ErrConflict):
		return fmt.Errorf("%s: %w", operation, ErrConflict)

	case errors.Is(err, repository.ErrInvalidReference):
		return fmt.Errorf("%s: %w", operation, ErrInvalidReference)

	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}
