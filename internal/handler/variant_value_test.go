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

type fakeVariantValueService struct {
	createFn func(
		context.Context,
		service.CreateVariantValueInput,
	) (domain.VariantValue, error)

	getFn func(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (domain.VariantValue, error)

	listFn func(
		context.Context,
		service.ListVariantValuesInput,
	) (service.Page[domain.VariantValue], error)

	listByVariantFn func(
		context.Context,
		uuid.UUID,
		service.PageInput,
	) (service.Page[domain.VariantValue], error)

	updateFn func(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		service.UpdateVariantValueInput,
	) (domain.VariantValue, error)
}

func (f *fakeVariantValueService) Create(
	ctx context.Context,
	input service.CreateVariantValueInput,
) (domain.VariantValue, error) {
	if f.createFn == nil {
		return domain.VariantValue{}, nil
	}

	return f.createFn(ctx, input)
}

func (f *fakeVariantValueService) GetByVariantAndAttribute(
	ctx context.Context,
	variantID uuid.UUID,
	attributeID uuid.UUID,
) (domain.VariantValue, error) {
	if f.getFn == nil {
		return domain.VariantValue{}, nil
	}

	return f.getFn(ctx, variantID, attributeID)
}

func (f *fakeVariantValueService) List(
	ctx context.Context,
	input service.ListVariantValuesInput,
) (service.Page[domain.VariantValue], error) {
	if f.listFn == nil {
		return service.Page[domain.VariantValue]{}, nil
	}

	return f.listFn(ctx, input)
}

func (f *fakeVariantValueService) ListByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
	page service.PageInput,
) (service.Page[domain.VariantValue], error) {
	if f.listByVariantFn == nil {
		return service.Page[domain.VariantValue]{}, nil
	}

	return f.listByVariantFn(ctx, variantID, page)
}

func (f *fakeVariantValueService) Update(
	ctx context.Context,
	variantID uuid.UUID,
	attributeID uuid.UUID,
	input service.UpdateVariantValueInput,
) (domain.VariantValue, error) {
	if f.updateFn == nil {
		return domain.VariantValue{}, nil
	}

	return f.updateFn(
		ctx,
		variantID,
		attributeID,
		input,
	)
}

func TestVariantValueHandlerCreate(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	variantID := uuid.New()
	attributeID := uuid.New()
	valueID := uuid.New()

	fake := &fakeVariantValueService{
		createFn: func(
			_ context.Context,
			input service.CreateVariantValueInput,
		) (domain.VariantValue, error) {
			return domain.VariantValue{
				VariantID:        input.VariantID,
				AttributeID:      input.AttributeID,
				AttributeValueID: input.AttributeValueID,
			}, nil
		},
	}

	h := NewVariantValueHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/variant-values", h.Create)

	body := []byte(
		`{"variant_id":"` +
			variantID.String() +
			`","attribute_id":"` +
			attributeID.String() +
			`","attribute_value_id":"` +
			valueID.String() +
			`"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/variant-values",
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

func TestVariantValueHandlerGet(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	variantID := uuid.New()
	attributeID := uuid.New()

	fake := &fakeVariantValueService{
		getFn: func(
			context.Context,
			uuid.UUID,
			uuid.UUID,
		) (domain.VariantValue, error) {
			return domain.VariantValue{
				VariantID:   variantID,
				AttributeID: attributeID,
			}, nil
		},
	}

	h := NewVariantValueHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET(
		"/variant-values/:variant_id/:attribute_id",
		h.Get,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/variant-values/"+
			variantID.String()+"/"+
			attributeID.String(),
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

func TestVariantValueHandlerInvalidUUID(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	h := NewVariantValueHandler(
		&fakeVariantValueService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET(
		"/variant-values/:variant_id/:attribute_id",
		h.Get,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/variant-values/not-a-uuid/"+
			uuid.New().String(),
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

func TestVariantValueHandlerConflict(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakeVariantValueService{
		createFn: func(
			context.Context,
			service.CreateVariantValueInput,
		) (domain.VariantValue, error) {
			return domain.VariantValue{},
				service.ErrConflict
		},
	}

	h := NewVariantValueHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/variant-values", h.Create)

	body := []byte(
		`{"variant_id":"` +
			uuid.New().String() +
			`","attribute_id":"` +
			uuid.New().String() +
			`","attribute_value_id":"` +
			uuid.New().String() +
			`"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/variant-values",
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
