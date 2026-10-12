package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type fakeBrandService struct {
	createFn    func(context.Context, service.CreateBrandInput) (domain.Brand, error)
	getByIDFn   func(context.Context, uuid.UUID) (domain.Brand, error)
	listFn      func(context.Context) ([]domain.Brand, error)
	updateFn    func(context.Context, uuid.UUID, service.UpdateBrandInput) (domain.Brand, error)
	setActiveFn func(context.Context, uuid.UUID, bool) (domain.Brand, error)
}

func (f *fakeBrandService) Create(
	ctx context.Context,
	input service.CreateBrandInput,
) (domain.Brand, error) {
	if f.createFn == nil {
		return domain.Brand{}, nil
	}

	return f.createFn(ctx, input)
}

func (f *fakeBrandService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Brand, error) {
	if f.getByIDFn == nil {
		return domain.Brand{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeBrandService) List(
	ctx context.Context,
) ([]domain.Brand, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakeBrandService) Update(
	ctx context.Context,
	id uuid.UUID,
	input service.UpdateBrandInput,
) (domain.Brand, error) {
	if f.updateFn == nil {
		return domain.Brand{}, nil
	}

	return f.updateFn(ctx, id, input)
}

func (f *fakeBrandService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Brand, error) {
	if f.setActiveFn == nil {
		return domain.Brand{}, nil
	}

	return f.setActiveFn(ctx, id, isActive)
}

func TestBrandHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	fakeService := &fakeBrandService{
		createFn: func(
			_ context.Context,
			input service.CreateBrandInput,
		) (domain.Brand, error) {
			if input.Name != "Nike" {
				t.Fatalf("Name = %q, want Nike", input.Name)
			}

			if input.Code != "NIKE" {
				t.Fatalf("Code = %q, want NIKE", input.Code)
			}

			return domain.Brand{
				ID:       id,
				Name:     input.Name,
				Code:     input.Code,
				IsActive: true,
			}, nil
		},
	}

	h := NewBrandHandler(
		fakeService,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/brands", h.Create)

	body := []byte(`{
		"name": "Nike",
		"code": "NIKE"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/brands",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusCreated,
		)
	}
}

func TestBrandHandlerCreateInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewBrandHandler(
		&fakeBrandService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/brands", h.Create)

	req := httptest.NewRequest(
		http.MethodPost,
		"/brands",
		bytes.NewBufferString(`{"name":`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
}

func TestBrandHandlerGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	fakeService := &fakeBrandService{
		getByIDFn: func(
			_ context.Context,
			gotID uuid.UUID,
		) (domain.Brand, error) {
			if gotID != id {
				t.Fatalf("ID = %s, want %s", gotID, id)
			}

			return domain.Brand{
				ID:   id,
				Name: "Nike",
				Code: "NIKE",
			}, nil
		},
	}

	h := NewBrandHandler(
		fakeService,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/brands/:id", h.GetByID)

	req := httptest.NewRequest(
		http.MethodGet,
		"/brands/"+id.String(),
		nil,
	)

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusOK,
		)
	}
}

func TestBrandHandlerGetByIDInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewBrandHandler(
		&fakeBrandService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/brands/:id", h.GetByID)

	req := httptest.NewRequest(
		http.MethodGet,
		"/brands/not-a-uuid",
		nil,
	)

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
}

func TestBrandHandlerGetByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fakeService := &fakeBrandService{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.Brand, error) {
			return domain.Brand{}, service.ErrNotFound
		},
	}

	h := NewBrandHandler(
		fakeService,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/brands/:id", h.GetByID)

	req := httptest.NewRequest(
		http.MethodGet,
		"/brands/"+uuid.NewString(),
		nil,
	)

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusNotFound,
		)
	}
}

func TestBrandHandlerCreateConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fakeService := &fakeBrandService{
		createFn: func(
			context.Context,
			service.CreateBrandInput,
		) (domain.Brand, error) {
			return domain.Brand{}, service.ErrConflict
		},
	}

	h := NewBrandHandler(
		fakeService,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/brands", h.Create)

	body := []byte(`{
		"name": "Nike",
		"code": "NIKE"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/brands",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusConflict,
		)
	}
}
