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

type fakeProductVariantService struct {
	createFn func(
		context.Context,
		service.CreateProductVariantInput,
	) (domain.ProductVariant, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (domain.ProductVariant, error)

	listFn func(
		context.Context,
		service.ListProductVariantsInput,
	) (service.Page[domain.ProductVariant], error)

	listByProductIDFn func(
		context.Context,
		uuid.UUID,
		service.PageInput,
	) (service.Page[domain.ProductVariant], error)

	listByIDsFn func(
		context.Context,
		[]uuid.UUID,
	) ([]domain.ProductVariant, error)

	updateFn func(
		context.Context,
		uuid.UUID,
		service.UpdateProductVariantInput,
	) (domain.ProductVariant, error)

	setActiveFn func(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.ProductVariant, error)
}

func (f *fakeProductVariantService) Create(
	ctx context.Context,
	input service.CreateProductVariantInput,
) (domain.ProductVariant, error) {
	if f.createFn == nil {
		return domain.ProductVariant{}, nil
	}

	return f.createFn(ctx, input)
}

func (f *fakeProductVariantService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.ProductVariant, error) {
	if f.getByIDFn == nil {
		return domain.ProductVariant{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeProductVariantService) List(
	ctx context.Context,
	input service.ListProductVariantsInput,
) (service.Page[domain.ProductVariant], error) {
	if f.listFn == nil {
		return service.Page[domain.ProductVariant]{}, nil
	}

	return f.listFn(ctx, input)
}

func (f *fakeProductVariantService) ListByProductID(
	ctx context.Context,
	productID uuid.UUID,
	page service.PageInput,
) (service.Page[domain.ProductVariant], error) {
	if f.listByProductIDFn == nil {
		return service.Page[domain.ProductVariant]{}, nil
	}

	return f.listByProductIDFn(ctx, productID, page)
}

func (f *fakeProductVariantService) ListByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]domain.ProductVariant, error) {
	if f.listByIDsFn == nil {
		return nil, nil
	}

	return f.listByIDsFn(ctx, ids)
}

func (f *fakeProductVariantService) Update(
	ctx context.Context,
	id uuid.UUID,
	input service.UpdateProductVariantInput,
) (domain.ProductVariant, error) {
	if f.updateFn == nil {
		return domain.ProductVariant{}, nil
	}

	return f.updateFn(ctx, id, input)
}

func (f *fakeProductVariantService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	active bool,
) (domain.ProductVariant, error) {
	if f.setActiveFn == nil {
		return domain.ProductVariant{}, nil
	}

	return f.setActiveFn(ctx, id, active)
}

func TestProductVariantHandlerCreate(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()
	productID := uuid.New()
	unitID := uuid.New()

	fake := &fakeProductVariantService{
		createFn: func(
			_ context.Context,
			input service.CreateProductVariantInput,
		) (domain.ProductVariant, error) {
			return domain.ProductVariant{
				ID:        id,
				ProductID: input.ProductID,
				SKU:       input.SKU,
				UnitID:    input.UnitID,
				IsActive:  true,
			}, nil
		},
	}

	h := NewProductVariantHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/product-variants", h.Create)

	body := []byte(
		`{"product_id":"` + productID.String() + `","sku":"SHIRT-BLUE-M","unit_id":"` + unitID.String() + `"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/product-variants",
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

func TestProductVariantHandlerGetByID(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	fake := &fakeProductVariantService{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.ProductVariant, error) {
			return domain.ProductVariant{
				ID:       id,
				SKU:      "SHIRT-BLUE-M",
				IsActive: true,
			}, nil
		},
	}

	h := NewProductVariantHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/product-variants/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/product-variants/"+id.String(),
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

func TestProductVariantHandlerList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fake := &fakeProductVariantService{
		listFn: func(
			_ context.Context,
			input service.ListProductVariantsInput,
		) (service.Page[domain.ProductVariant], error) {
			if input.Page.Limit != 10 {
				t.Fatalf("expected limit 10, got %d", input.Page.Limit)
			}
			return service.Page[domain.ProductVariant]{
				Items: []domain.ProductVariant{
					{ID: uuid.New(), SKU: "VAR-1"},
				},
			}, nil
		},
	}

	h := NewProductVariantHandler(fake, zerolog.Nop())
	r := gin.New()
	r.GET("/product-variants", h.List)

	req := httptest.NewRequest(http.MethodGet, "/product-variants?limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestProductVariantHandlerBatchList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id1 := uuid.New()
	id2 := uuid.New()

	fake := &fakeProductVariantService{
		listByIDsFn: func(
			_ context.Context,
			ids []uuid.UUID,
		) ([]domain.ProductVariant, error) {
			if len(ids) != 2 {
				t.Fatalf("expected 2 ids, got %d", len(ids))
			}
			return []domain.ProductVariant{
				{ID: id1, SKU: "VAR-1"},
				{ID: id2, SKU: "VAR-2"},
			}, nil
		},
	}

	h := NewProductVariantHandler(fake, zerolog.Nop())
	r := gin.New()
	r.GET("/product-variants", h.List)

	req := httptest.NewRequest(http.MethodGet, "/product-variants?ids="+id1.String()+","+id2.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
