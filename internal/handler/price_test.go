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

type fakePriceService struct {
	createFn func(
		context.Context,
		service.CreatePriceInput,
	) (domain.Price, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (domain.Price, error)

	listFn func(
		context.Context,
	) ([]domain.Price, error)

	listByPriceListFn func(
		context.Context,
		uuid.UUID,
	) ([]domain.Price, error)

	listByVariantFn func(
		context.Context,
		uuid.UUID,
	) ([]domain.Price, error)

	updateFn func(
		context.Context,
		uuid.UUID,
		service.UpdatePriceInput,
	) (domain.Price, error)
}

func (f *fakePriceService) Create(
	ctx context.Context,
	input service.CreatePriceInput,
) (domain.Price, error) {
	if f.createFn == nil {
		return domain.Price{}, nil
	}

	return f.createFn(ctx, input)
}

func (f *fakePriceService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Price, error) {
	if f.getByIDFn == nil {
		return domain.Price{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakePriceService) List(
	ctx context.Context,
) ([]domain.Price, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakePriceService) ListByPriceListID(
	ctx context.Context,
	id uuid.UUID,
) ([]domain.Price, error) {
	if f.listByPriceListFn == nil {
		return nil, nil
	}

	return f.listByPriceListFn(ctx, id)
}

func (f *fakePriceService) ListByVariantID(
	ctx context.Context,
	id uuid.UUID,
) ([]domain.Price, error) {
	if f.listByVariantFn == nil {
		return nil, nil
	}

	return f.listByVariantFn(ctx, id)
}

func (f *fakePriceService) Update(
	ctx context.Context,
	id uuid.UUID,
	input service.UpdatePriceInput,
) (domain.Price, error) {
	if f.updateFn == nil {
		return domain.Price{}, nil
	}

	return f.updateFn(
		ctx,
		id,
		input,
	)
}

func TestPriceHandlerCreate(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	priceListID := uuid.New()
	variantID := uuid.New()
	id := uuid.New()

	fake := &fakePriceService{
		createFn: func(
			_ context.Context,
			input service.CreatePriceInput,
		) (domain.Price, error) {
			return domain.Price{
				ID:          id,
				PriceListID: input.PriceListID,
				VariantID:   input.VariantID,
				UnitPrice:   input.UnitPrice,
			}, nil
		},
	}

	h := NewPriceHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/prices", h.Create)

	body := []byte(
		`{"price_list_id":"` +
			priceListID.String() +
			`","variant_id":"` +
			variantID.String() +
			`","unit_price":"850.00"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/prices",
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

func TestPriceHandlerInvalidJSON(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	h := NewPriceHandler(
		&fakePriceService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/prices", h.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/prices",
		bytes.NewBufferString(`{"unit_price":`),
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

func TestPriceHandlerInvalidUUID(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	h := NewPriceHandler(
		&fakePriceService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/prices/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/prices/not-a-uuid",
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

func TestPriceHandlerNotFound(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakePriceService{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.Price, error) {
			return domain.Price{},
				service.ErrNotFound
		},
	}

	h := NewPriceHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/prices/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/prices/"+uuid.New().String(),
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

func TestPriceHandlerConflict(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakePriceService{
		createFn: func(
			context.Context,
			service.CreatePriceInput,
		) (domain.Price, error) {
			return domain.Price{},
				service.ErrConflict
		},
	}

	h := NewPriceHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/prices", h.Create)

	body := []byte(
		`{"price_list_id":"` +
			uuid.New().String() +
			`","variant_id":"` +
			uuid.New().String() +
			`","unit_price":"850.00"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/prices",
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

func TestPriceHandlerInvalidReference(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakePriceService{
		createFn: func(
			context.Context,
			service.CreatePriceInput,
		) (domain.Price, error) {
			return domain.Price{},
				service.ErrInvalidReference
		},
	}

	h := NewPriceHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/prices", h.Create)

	body := []byte(
		`{"price_list_id":"` +
			uuid.New().String() +
			`","variant_id":"` +
			uuid.New().String() +
			`","unit_price":"850.00"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/prices",
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
