package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultPort                 = "8081"
	defaultGinMode              = "release"
	defaultDBMaxConns           = 10
	defaultDBMinConns           = 2
	defaultDBMaxConnLifetime    = 30 * time.Minute
	defaultDBMaxConnIdleTime    = 5 * time.Minute
	defaultDBHealthCheckPeriod  = time.Minute
	defaultDBStatementTimeout   = 5 * time.Second
	defaultMaxRequestBodyBytes  = 1 << 20
	defaultSlowRequestThreshold = 500 * time.Millisecond
	defaultReferenceCacheTTL    = 45 * time.Second
)

type DatabaseConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	StatementTimeout  time.Duration
}

type HTTPConfig struct {
	MaxRequestBodyBytes  int64
	SlowRequestThreshold time.Duration
}

type CacheConfig struct {
	ReferenceTTL time.Duration
}

type Config struct {
	JWTPublicKeyPath string
	JWTIssuer        string
	JWTAudience      string
	Port             string
	DatabaseURL      string
	GinMode          string
	Database         DatabaseConfig
	HTTP             HTTPConfig
	Cache            CacheConfig
}

func Load() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	dbMaxConns, err := getEnvInt(
		"DB_MAX_CONNS",
		defaultDBMaxConns,
	)
	if err != nil {
		return nil, err
	}

	dbMinConns, err := getEnvInt(
		"DB_MIN_CONNS",
		defaultDBMinConns,
	)
	if err != nil {
		return nil, err
	}

	if dbMaxConns <= 0 {
		return nil, errors.New("DB_MAX_CONNS must be greater than zero")
	}

	if dbMinConns < 0 {
		return nil, errors.New("DB_MIN_CONNS cannot be negative")
	}

	if dbMinConns > dbMaxConns {
		return nil, errors.New(
			"DB_MIN_CONNS cannot be greater than DB_MAX_CONNS",
		)
	}

	maxConnLifetime, err := getEnvDuration(
		"DB_MAX_CONN_LIFETIME",
		defaultDBMaxConnLifetime,
	)
	if err != nil {
		return nil, err
	}

	maxConnIdleTime, err := getEnvDuration(
		"DB_MAX_CONN_IDLE_TIME",
		defaultDBMaxConnIdleTime,
	)
	if err != nil {
		return nil, err
	}

	healthCheckPeriod, err := getEnvDuration(
		"DB_HEALTH_CHECK_PERIOD",
		defaultDBHealthCheckPeriod,
	)
	if err != nil {
		return nil, err
	}

	statementTimeout, err := getEnvDuration(
		"DB_STATEMENT_TIMEOUT",
		defaultDBStatementTimeout,
	)
	if err != nil {
		return nil, err
	}

	maxRequestBodyBytes, err := getEnvInt(
		"HTTP_MAX_BODY_BYTES",
		defaultMaxRequestBodyBytes,
	)
	if err != nil {
		return nil, err
	}

	if maxRequestBodyBytes <= 0 {
		return nil, errors.New(
			"HTTP_MAX_BODY_BYTES must be greater than zero",
		)
	}

	slowRequestThreshold, err := getEnvDuration(
		"HTTP_SLOW_REQUEST_THRESHOLD",
		defaultSlowRequestThreshold,
	)
	if err != nil {
		return nil, err
	}

	cacheTTL, err := getEnvDuration(
		"REFERENCE_CACHE_TTL",
		defaultReferenceCacheTTL,
	)
	if err != nil {
		return nil, err
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = defaultGinMode
	}

	return &Config{
		JWTPublicKeyPath: os.Getenv("JWT_PUBLIC_KEY_PATH"),
		JWTIssuer:        os.Getenv("JWT_ISSUER"),
		JWTAudience:      os.Getenv("JWT_AUDIENCE"),
		Port:             port,
		DatabaseURL:      databaseURL,
		GinMode:          ginMode,

		Database: DatabaseConfig{
			MaxConns:          int32(dbMaxConns),
			MinConns:          int32(dbMinConns),
			MaxConnLifetime:   maxConnLifetime,
			MaxConnIdleTime:   maxConnIdleTime,
			HealthCheckPeriod: healthCheckPeriod,
			StatementTimeout:  statementTimeout,
		},

		HTTP: HTTPConfig{
			MaxRequestBodyBytes:  int64(maxRequestBodyBytes),
			SlowRequestThreshold: slowRequestThreshold,
		},

		Cache: CacheConfig{
			ReferenceTTL: cacheTTL,
		},
	}, nil
}

func getEnvInt(
	name string,
	defaultValue int,
) (int, error) {
	value := os.Getenv(name)

	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid %s: %w",
			name,
			err,
		)
	}

	return parsed, nil
}

func getEnvDuration(
	name string,
	defaultValue time.Duration,
) (time.Duration, error) {
	value := os.Getenv(name)

	if value == "" {
		return defaultValue, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid %s: %w",
			name,
			err,
		)
	}

	return parsed, nil
}
