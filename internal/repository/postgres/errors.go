package postgres

import (
	"errors"
	"fmt"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapDBError(
	operation string,
	err error,
) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf(
			"%s: %w",
			operation,
			repository.ErrNotFound,
		)
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf(
				"%s: %w: %w",
				operation,
				repository.ErrConflict,
				err,
			)

		case "23503":
			return fmt.Errorf(
				"%s: %w: %w",
				operation,
				repository.ErrInvalidReference,
				err,
			)
		}
	}

	return fmt.Errorf(
		"%s: %w",
		operation,
		err,
	)
}
