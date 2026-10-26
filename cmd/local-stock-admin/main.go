// Local operator utility for explicitly confirmed physical definitions.
package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/strictjson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type selection struct {
	Unit         uuid.UUID   `json:"unit_id"`
	Variants     []uuid.UUID `json:"variant_ids"`
	Confirmation string      `json:"confirmation"`
}
type definitions interface {
	ClassifyPieceUnit(context.Context, uuid.UUID) error
	EnableStock(context.Context, uuid.UUID, uuid.UUID) error
}

var failed = errors.New("local classification stopped; inspect definitions before retry")

func apply(ctx context.Context, s definitions, p selection) error {
	if p.Unit == uuid.Nil || len(p.Variants) < 1 || len(p.Variants) > 100 || p.Confirmation != "one base unit is one physical piece for every variant using this unit" {
		return failed
	}
	seen := map[uuid.UUID]bool{}
	for _, v := range p.Variants {
		if v == uuid.Nil || seen[v] {
			return failed
		}
		seen[v] = true
	}
	if s.ClassifyPieceUnit(ctx, p.Unit) != nil {
		return failed
	}
	for _, v := range p.Variants {
		if s.EnableStock(ctx, v, p.Unit) != nil {
			return failed
		}
	}
	return nil
}
func decodeSelection(r io.Reader) (selection, error) {
	var p selection
	b, e := io.ReadAll(io.LimitReader(r, (64<<10)+1))
	if e != nil || len(b) > 64<<10 || strictjson.Decode(b, &p) != nil {
		return p, failed
	}
	return p, nil
}

func localHost(host string) bool {
	return host == "postgres" || host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func run() error {
	path := flag.String("selection-file", "", "confirmed operator selection")
	ack := flag.String("local-database", "", "explicit local target database")
	host := flag.String("local-host", "", "explicit local target host (postgres or loopback)")
	flag.Parse()
	if *ack == "" || !filepath.IsAbs(*path) || !localHost(*host) {
		return failed
	}
	f, e := os.Open(*path)
	if e != nil {
		return failed
	}
	defer f.Close()
	p, e := decodeSelection(f)
	if e != nil {
		return failed
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, e := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if e != nil || cfg.ConnConfig.Database != *ack || cfg.ConnConfig.Host != *host {
		return failed
	}
	for _, fallback := range cfg.ConnConfig.Fallbacks {
		if fallback.Host != *host {
			return failed
		}
	}
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		return failed
	}
	defer pool.Close()
	s, e := service.NewStockReferenceService(postgres.NewStockReferenceRepository(pool))
	if e != nil {
		return failed
	}
	return apply(ctx, s, p)
}
func main() {
	if run() != nil {
		log := zerolog.New(os.Stderr)
		log.Error().Str("component", "local_stock_admin").Str("operation", "classify").Str("reason", "reconciliation_required").Msg("Classification stopped; inspect selected definitions")
		os.Exit(1)
	}
	log := zerolog.New(os.Stdout)
	log.Info().Str("component", "local_stock_admin").Str("operation", "classify").Str("reason", "completed").Msg("Confirmed physical definitions applied; no stock posted")
}
