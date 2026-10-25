package config

import "testing"

func TestStockReferenceTLSConfiguration(t *testing.T) {
	for _, key := range []string{"CATALOG_STOCK_REFERENCES_ENABLED", "TLS_CERT_PATH", "TLS_KEY_PATH", "DB_MAX_CONNS", "DB_MIN_CONNS", "DB_MAX_CONN_LIFETIME", "DB_MAX_CONN_IDLE_TIME", "DB_HEALTH_CHECK_PERIOD", "DB_STATEMENT_TIMEOUT", "HTTP_MAX_BODY_BYTES", "HTTP_SLOW_REQUEST_THRESHOLD", "REFERENCE_CACHE_TTL"} {
		t.Setenv(key, "")
	}
	t.Setenv("DATABASE_URL", "synthetic configuration only")
	if cfg, err := Load(); err != nil || cfg.StockReferencesEnabled {
		t.Fatal("disabled default changed")
	}
	t.Setenv("CATALOG_STOCK_REFERENCES_ENABLED", "true")
	if _, err := Load(); err == nil {
		t.Fatal("enabled plaintext accepted")
	}
	t.Setenv("TLS_CERT_PATH", "synthetic-cert.pem")
	if _, err := Load(); err == nil {
		t.Fatal("unpaired TLS accepted")
	}
	t.Setenv("TLS_KEY_PATH", "synthetic-key.pem")
	if cfg, err := Load(); err != nil || !cfg.StockReferencesEnabled {
		t.Fatal("paired configuration rejected")
	}
	t.Setenv("CATALOG_STOCK_REFERENCES_ENABLED", "invalid")
	if _, err := Load(); err == nil {
		t.Fatal("invalid enable flag accepted")
	}
}
