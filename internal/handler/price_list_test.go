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

type fakePriceListService struct {
	createFn func(
		context.Context,
		service.CreatePriceListInput,
	) (domain.PriceList, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (domain.PriceList, error)

	listFn func(
		context.Context,
	) ([]domain.PriceList, error)

	updateFn func(
		context.Context,
		uuid.UUID,
		service.UpdatePriceListInput,
	) (domain.PriceList, error)

	setActiveFn func(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.PriceList, error)
}

func (f *fakePriceListService) Create(
	ctx context.Context,
	input service.CreatePriceListInput,
) (domain.PriceList, error) {
	if f.createFn == nil {
		return domain.PriceList{}, nil
	}

	return f.createFn(ctx, input)
}

func (f *fakePriceListService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.PriceList, error) {
	if f.getByIDFn == nil {
		return domain.PriceList{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakePriceListService) List(
	ctx context.Context,
) ([]domain.PriceList, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakePriceListService) Update(
	ctx context.Context,
	id uuid.UUID,
	input service.UpdatePriceListInput,
) (domain.PriceList, error) {
	if f.updateFn == nil {
		return domain.PriceList{}, nil
	}

	return f.updateFn(
		ctx,
		id,
		input,
	)
}

func (f *fakePriceListService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	active bool,
) (domain.PriceList, error) {
	if f.setActiveFn == nil {
		return domain.PriceList{}, nil
	}

	return f.setActiveFn(
		ctx,
		id,
		active,
	)
}

func TestPriceListHandlerCreate(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	fake := &fakePriceListService{
		createFn: func(
			_ context.Context,
			input service.CreatePriceListInput,
		) (domain.PriceList, error) {
			return domain.PriceList{
				ID:           id,
				Name:         input.Name,
				Code:         input.Code,
				CurrencyCode: input.CurrencyCode,
				IsActive:     true,
			}, nil
		},
	}

	h := NewPriceListHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/price-lists", h.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/price-lists",
		bytes.NewBufferString(
			`{"name":"Wholesale","code":"WHOLESALE","currency_code":"INR"}`,
		),
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

func TestPriceListHandlerInvalidJSON(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	h := NewPriceListHandler(
		&fakePriceListService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/price-lists", h.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/price-lists",
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

func TestPriceListHandlerInvalidUUID(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	h := NewPriceListHandler(
		&fakePriceListService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/price-lists/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/price-lists/not-a-uuid",
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

func TestPriceListHandlerNotFound(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakePriceListService{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.PriceList, error) {
			return domain.PriceList{},
				service.ErrNotFound
		},
	}

	h := NewPriceListHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/price-lists/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/price-lists/"+uuid.New().String(),
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

func TestPriceListHandlerConflict(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakePriceListService{
		createFn: func(
			context.Context,
			service.CreatePriceListInput,
		) (domain.PriceList, error) {
			return domain.PriceList{},
				service.ErrConflict
		},
	}

	h := NewPriceListHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/price-lists", h.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/price-lists",
		bytes.NewBufferString(
			`{"name":"Wholesale","code":"WHOLESALE","currency_code":"INR"}`,
		),
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

func TestPriceListHandlerSetActive(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	fake := &fakePriceListService{
		setActiveFn: func(
			_ context.Context,
			gotID uuid.UUID,
			active bool,
		) (domain.PriceList, error) {
			return domain.PriceList{
				ID:       gotID,
				IsActive: active,
			}, nil
		},
	}

	h := NewPriceListHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.PATCH("/price-lists/:id/status", h.SetActive)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/price-lists/"+id.String()+"/status",
		bytes.NewBufferString(
			`{"is_active":false}`,
		),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
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
