package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type fakeCategoryService struct {
	createFn  func(context.Context, service.CreateCategoryInput) (domain.Category, error)
	getByIDFn func(context.Context, uuid.UUID) (domain.Category, error)
}

func (f *fakeCategoryService) Create(
	ctx context.Context,
	input service.CreateCategoryInput,
) (domain.Category, error) {
	if f.createFn != nil {
		return f.createFn(ctx, input)
	}

	return domain.Category{}, nil
}

func (f *fakeCategoryService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Category, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}

	return domain.Category{}, nil
}

func (f *fakeCategoryService) GetByCode(
	context.Context,
	string,
) (domain.Category, error) {
	return domain.Category{}, nil
}

func (f *fakeCategoryService) List(
	context.Context,
) ([]domain.Category, error) {
	return nil, nil
}

func (f *fakeCategoryService) Update(
	context.Context,
	uuid.UUID,
	service.UpdateCategoryInput,
) (domain.Category, error) {
	return domain.Category{}, nil
}

func (f *fakeCategoryService) SetActive(
	context.Context,
	uuid.UUID,
	bool,
) (domain.Category, error) {
	return domain.Category{}, nil
}

func TestCategoryHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	categoryID := uuid.New()

	fakeService := &fakeCategoryService{
		createFn: func(
			_ context.Context,
			input service.CreateCategoryInput,
		) (domain.Category, error) {
			if input.Name != "Shirts" {
				t.Fatalf("Name = %q, want Shirts", input.Name)
			}

			if input.Code != "SHIRTS" {
				t.Fatalf("Code = %q, want SHIRTS", input.Code)
			}

			return domain.Category{
				ID:       categoryID,
				Name:     input.Name,
				Code:     input.Code,
				IsActive: true,
			}, nil
		},
	}

	handler := NewCategoryHandler(
		fakeService,
		zerolog.Nop(),
	)

	router := gin.New()
	router.POST("/categories", handler.Create)

	body := []byte(`{
		"name": "Shirts",
		"code": "SHIRTS"
	}`)

	request := httptest.NewRequest(
		http.MethodPost,
		"/categories",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusCreated,
		)
	}
}

func TestCategoryHandlerCreateInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewCategoryHandler(
		&fakeCategoryService{},
		zerolog.Nop(),
	)

	router := gin.New()
	router.POST("/categories", handler.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/categories",
		bytes.NewBufferString(`{"name":`),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusBadRequest,
		)
	}
}

func TestCategoryHandlerGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	fakeService := &fakeCategoryService{
		getByIDFn: func(
			_ context.Context,
			gotID uuid.UUID,
		) (domain.Category, error) {
			if gotID != id {
				t.Fatalf("ID = %s, want %s", gotID, id)
			}

			return domain.Category{
				ID:   id,
				Name: "Shirts",
				Code: "SHIRTS",
			}, nil
		},
	}

	handler := NewCategoryHandler(
		fakeService,
		zerolog.Nop(),
	)

	router := gin.New()
	router.GET("/categories/:id", handler.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/categories/"+id.String(),
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusOK,
		)
	}
}

func TestCategoryHandlerGetByIDInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewCategoryHandler(
		&fakeCategoryService{},
		zerolog.Nop(),
	)

	router := gin.New()
	router.GET("/categories/:id", handler.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/categories/not-a-uuid",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusBadRequest,
		)
	}
}

func TestCategoryHandlerGetByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fakeService := &fakeCategoryService{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.Category, error) {
			return domain.Category{}, service.ErrNotFound
		},
	}

	handler := NewCategoryHandler(
		fakeService,
		zerolog.Nop(),
	)

	router := gin.New()
	router.GET("/categories/:id", handler.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/categories/"+uuid.NewString(),
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusNotFound,
		)
	}
}

func TestCategoryHandlerCreateConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fakeService := &fakeCategoryService{
		createFn: func(
			context.Context,
			service.CreateCategoryInput,
		) (domain.Category, error) {
			return domain.Category{}, errors.New(
				"create: " + service.ErrConflict.Error(),
			)
		},
	}

	handler := NewCategoryHandler(
		fakeService,
		zerolog.Nop(),
	)

	router := gin.New()
	router.POST("/categories", handler.Create)

	body := []byte(`{
		"name": "Shirts",
		"code": "SHIRTS"
	}`)

	request := httptest.NewRequest(
		http.MethodPost,
		"/categories",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusInternalServerError,
		)
	}
}
