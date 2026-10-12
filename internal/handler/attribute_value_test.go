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

type fakeAttributeValueService struct {
	createFn func(
		context.Context,
		service.CreateAttributeValueInput,
	) (domain.AttributeValue, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (domain.AttributeValue, error)

	listFn func(
		context.Context,
	) ([]domain.AttributeValue, error)

	listByAttributeIDFn func(
		context.Context,
		uuid.UUID,
	) ([]domain.AttributeValue, error)

	updateFn func(
		context.Context,
		uuid.UUID,
		service.UpdateAttributeValueInput,
	) (domain.AttributeValue, error)
}

func (f *fakeAttributeValueService) Create(
	ctx context.Context,
	input service.CreateAttributeValueInput,
) (domain.AttributeValue, error) {
	if f.createFn == nil {
		return domain.AttributeValue{}, nil
	}

	return f.createFn(ctx, input)
}

func (f *fakeAttributeValueService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.AttributeValue, error) {
	if f.getByIDFn == nil {
		return domain.AttributeValue{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeAttributeValueService) List(
	ctx context.Context,
) ([]domain.AttributeValue, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakeAttributeValueService) ListByAttributeID(
	ctx context.Context,
	attributeID uuid.UUID,
) ([]domain.AttributeValue, error) {
	if f.listByAttributeIDFn == nil {
		return nil, nil
	}

	return f.listByAttributeIDFn(
		ctx,
		attributeID,
	)
}

func (f *fakeAttributeValueService) Update(
	ctx context.Context,
	id uuid.UUID,
	input service.UpdateAttributeValueInput,
) (domain.AttributeValue, error) {
	if f.updateFn == nil {
		return domain.AttributeValue{}, nil
	}

	return f.updateFn(
		ctx,
		id,
		input,
	)
}

func TestAttributeValueHandlerCreate(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	attributeID := uuid.New()
	id := uuid.New()

	fake := &fakeAttributeValueService{
		createFn: func(
			_ context.Context,
			input service.CreateAttributeValueInput,
		) (domain.AttributeValue, error) {
			return domain.AttributeValue{
				ID:          id,
				AttributeID: input.AttributeID,
				Value:       input.Value,
			}, nil
		},
	}

	h := NewAttributeValueHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/attribute-values", h.Create)

	body := []byte(
		`{"attribute_id":"` +
			attributeID.String() +
			`","value":"Blue"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/attribute-values",
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

func TestAttributeValueHandlerInvalidJSON(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	h := NewAttributeValueHandler(
		&fakeAttributeValueService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/attribute-values", h.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/attribute-values",
		bytes.NewBufferString(`{"value":`),
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

func TestAttributeValueHandlerGetByID(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	fake := &fakeAttributeValueService{
		getByIDFn: func(
			_ context.Context,
			gotID uuid.UUID,
		) (domain.AttributeValue, error) {
			if gotID != id {
				t.Fatal("unexpected ID")
			}

			return domain.AttributeValue{
				ID:    id,
				Value: "Blue",
			}, nil
		},
	}

	h := NewAttributeValueHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/attribute-values/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/attribute-values/"+id.String(),
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

func TestAttributeValueHandlerInvalidUUID(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	h := NewAttributeValueHandler(
		&fakeAttributeValueService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/attribute-values/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/attribute-values/not-a-uuid",
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

func TestAttributeValueHandlerNotFound(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakeAttributeValueService{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.AttributeValue, error) {
			return domain.AttributeValue{},
				service.ErrNotFound
		},
	}

	h := NewAttributeValueHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/attribute-values/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/attribute-values/"+uuid.New().String(),
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

func TestAttributeValueHandlerConflict(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakeAttributeValueService{
		createFn: func(
			context.Context,
			service.CreateAttributeValueInput,
		) (domain.AttributeValue, error) {
			return domain.AttributeValue{},
				service.ErrConflict
		},
	}

	h := NewAttributeValueHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/attribute-values", h.Create)

	body := []byte(
		`{"attribute_id":"` +
			uuid.New().String() +
			`","value":"Blue"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/attribute-values",
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

func TestAttributeValueHandlerInvalidReference(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakeAttributeValueService{
		createFn: func(
			context.Context,
			service.CreateAttributeValueInput,
		) (domain.AttributeValue, error) {
			return domain.AttributeValue{},
				service.ErrInvalidReference
		},
	}

	h := NewAttributeValueHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/attribute-values", h.Create)

	body := []byte(
		`{"attribute_id":"` +
			uuid.New().String() +
			`","value":"Blue"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/attribute-values",
		bytes.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code !=
		http.StatusUnprocessableEntity {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnprocessableEntity,
			response.Code,
		)
	}
}
