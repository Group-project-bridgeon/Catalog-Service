package app

import (
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestInvalidAuthConfigurationStopsBeforeDatabaseConnection(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://unreachable.invalid/catalog")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "")
	t.Setenv("JWT_ISSUER", "erp-auth")
	t.Setenv("JWT_AUDIENCE", "erp-api")
	err := Run(zerolog.Nop())
	if err == nil || !strings.Contains(err.Error(), "load JWT verifier") {
		t.Fatalf("expected startup JWT error before DB connection, got %v", err)
	}
}
