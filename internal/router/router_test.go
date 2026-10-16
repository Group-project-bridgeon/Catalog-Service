package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

type fakeDatabase struct {
	err error
}

func (f fakeDatabase) Ping(
	context.Context,
) error {
	return f.err
}

func TestHealth(t *testing.T) {
	log := zerolog.Nop()

	r := New(
		fakeDatabase{},
		log,
		1<<20,
		500*time.Millisecond,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestReadyWhenDatabaseAvailable(t *testing.T) {
	log := zerolog.Nop()

	r := New(
		fakeDatabase{},
		log,
		1<<20,
		500*time.Millisecond,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/ready",
		nil,
	)

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestReadyWhenDatabaseUnavailable(t *testing.T) {
	log := zerolog.Nop()

	r := New(
		fakeDatabase{
			err: errors.New("database unavailable"),
		},
		log,
		1<<20,
		500*time.Millisecond,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/ready",
		nil,
	)

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}
}
