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

type fakeAttributeService struct {
	createFn func(
		context.Context,
		service.CreateAttributeInput,
	) (domain.Attribute, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (domain.Attribute, error)

	listFn func(
		context.Context,
	) ([]domain.Attribute, error)

	updateFn func(
		context.Context,
		uuid.UUID,
		service.UpdateAttributeInput,
	) (domain.Attribute, error)
}

func (f *fakeAttributeService) Create(
	ctx context.Context,
	input service.CreateAttributeInput,
) (domain.Attribute, error) {
	if f.createFn == nil {
		return domain.Attribute{}, nil
	}

	return f.createFn(ctx, input)
}

func (f *fakeAttributeService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Attribute, error) {
	if f.getByIDFn == nil {
		return domain.Attribute{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeAttributeService) List(
	ctx context.Context,
) ([]domain.Attribute, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakeAttributeService) Update(
	ctx context.Context,
	id uuid.UUID,
	input service.UpdateAttributeInput,
) (domain.Attribute, error) {
	if f.updateFn == nil {
		return domain.Attribute{}, nil
	}

	return f.updateFn(
		ctx,
		id,
		input,
	)
}

func TestAttributeHandlerCreate(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	fake := &fakeAttributeService{
		createFn: func(
			_ context.Context,
			input service.CreateAttributeInput,
		) (domain.Attribute, error) {
			return domain.Attribute{
				ID:   id,
				Name: input.Name,
				Code: input.Code,
			}, nil
		},
	}

	h := NewAttributeHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/attributes", h.Create)

	body := []byte(
		`{"name":"Color","code":"COLOR"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/attributes",
		bytes.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusCreated,
			response.Code,
		)
	}
}

func TestAttributeHandlerInvalidJSON(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	h := NewAttributeHandler(
		&fakeAttributeService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/attributes", h.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/attributes",
		bytes.NewBufferString(`{"name":`),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestAttributeHandlerGetByID(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	fake := &fakeAttributeService{
		getByIDFn: func(
			_ context.Context,
			gotID uuid.UUID,
		) (domain.Attribute, error) {
			if gotID != id {
				t.Fatal("unexpected ID")
			}

			return domain.Attribute{
				ID:   id,
				Name: "Color",
				Code: "COLOR",
			}, nil
		},
	}

	h := NewAttributeHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/attributes/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/attributes/"+id.String(),
		nil,
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusOK,
			response.Code,
		)
	}
}

func TestAttributeHandlerInvalidUUID(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	h := NewAttributeHandler(
		&fakeAttributeService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/attributes/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/attributes/not-a-uuid",
		nil,
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestAttributeHandlerNotFound(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakeAttributeService{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.Attribute, error) {
			return domain.Attribute{},
				service.ErrNotFound
		},
	}

	h := NewAttributeHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/attributes/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/attributes/"+uuid.New().String(),
		nil,
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusNotFound,
			response.Code,
		)
	}
}

func TestAttributeHandlerConflict(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakeAttributeService{
		createFn: func(
			context.Context,
			service.CreateAttributeInput,
		) (domain.Attribute, error) {
			return domain.Attribute{},
				service.ErrConflict
		},
	}

	h := NewAttributeHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/attributes", h.Create)

	body := []byte(
		`{"name":"Color","code":"COLOR"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/attributes",
		bytes.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusConflict,
			response.Code,
		)
	}
}
