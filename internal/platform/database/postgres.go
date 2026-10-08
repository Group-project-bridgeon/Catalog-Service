package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	StatementTimeout  time.Duration
}

func NewPool(
	ctx context.Context,
	databaseURL string,
	settings PoolConfig,
) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"parse postgres config: %w",
			err,
		)
	}

	cfg.MaxConns = settings.MaxConns
	cfg.MinConns = settings.MinConns
	cfg.MaxConnLifetime = settings.MaxConnLifetime
	cfg.MaxConnIdleTime = settings.MaxConnIdleTime
	cfg.HealthCheckPeriod = settings.HealthCheckPeriod

	if settings.StatementTimeout > 0 {
		cfg.ConnConfig.RuntimeParams["statement_timeout"] =
			settings.StatementTimeout.String()
	}

	pool, err := pgxpool.NewWithConfig(
		ctx,
		cfg,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create postgres pool: %w",
			err,
		)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf(
			"ping postgres: %w",
			err,
		)
	}

	return pool, nil
}
