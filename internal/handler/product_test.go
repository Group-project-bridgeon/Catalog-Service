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

type fakeProductService struct {
	createFn    func(context.Context, service.CreateProductInput) (domain.Product, error)
	getByIDFn   func(context.Context, uuid.UUID) (domain.Product, error)
	listFn      func(context.Context, service.ListProductsInput) (service.Page[domain.Product], error)
	updateFn    func(context.Context, uuid.UUID, service.UpdateProductInput) (domain.Product, error)
	setActiveFn func(context.Context, uuid.UUID, bool) (domain.Product, error)
}

func (f *fakeProductService) Create(
	ctx context.Context,
	input service.CreateProductInput,
) (domain.Product, error) {
	if f.createFn == nil {
		return domain.Product{}, nil
	}

	return f.createFn(ctx, input)
}

func (f *fakeProductService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Product, error) {
	if f.getByIDFn == nil {
		return domain.Product{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeProductService) List(
	ctx context.Context,
	input service.ListProductsInput,
) (service.Page[domain.Product], error) {
	if f.listFn == nil {
		return service.Page[domain.Product]{}, nil
	}

	return f.listFn(ctx, input)
}

func (f *fakeProductService) Update(
	ctx context.Context,
	id uuid.UUID,
	input service.UpdateProductInput,
) (domain.Product, error) {
	if f.updateFn == nil {
		return domain.Product{}, nil
	}

	return f.updateFn(ctx, id, input)
}

func (f *fakeProductService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	active bool,
) (domain.Product, error) {
	if f.setActiveFn == nil {
		return domain.Product{}, nil
	}

	return f.setActiveFn(ctx, id, active)
}

func TestProductHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	categoryID := uuid.New()
	productID := uuid.New()

	fake := &fakeProductService{
		createFn: func(
			_ context.Context,
			input service.CreateProductInput,
		) (domain.Product, error) {
			return domain.Product{
				ID:          productID,
				CategoryID:  categoryID,
				Name:        input.Name,
				ProductCode: input.ProductCode,
				IsActive:    true,
			}, nil
		},
	}

	h := NewProductHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/products", h.Create)

	body := []byte(
		`{"category_id":"` +
			categoryID.String() +
			`","name":"Formal Shirt","product_code":"P001"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/products",
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
			"expected status %d, got %d",
			http.StatusCreated,
			response.Code,
		)
	}
}

func TestProductHandlerCreateInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProductHandler(
		&fakeProductService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/products", h.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/products",
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
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestProductHandlerGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()

	fake := &fakeProductService{
		getByIDFn: func(
			_ context.Context,
			gotID uuid.UUID,
		) (domain.Product, error) {
			if gotID != id {
				t.Fatal("unexpected product ID")
			}

			return domain.Product{
				ID:          id,
				CategoryID:  uuid.New(),
				Name:        "Product",
				ProductCode: "P001",
			}, nil
		},
	}

	h := NewProductHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/products/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products/"+id.String(),
		nil,
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			response.Code,
		)
	}
}

func TestProductHandlerGetByIDInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProductHandler(
		&fakeProductService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/products/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products/not-a-uuid",
		nil,
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestProductHandlerGetByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fake := &fakeProductService{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.Product, error) {
			return domain.Product{},
				service.ErrNotFound
		},
	}

	h := NewProductHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/products/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products/"+uuid.New().String(),
		nil,
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			response.Code,
		)
	}
}

func TestProductHandlerCreateConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	categoryID := uuid.New()

	fake := &fakeProductService{
		createFn: func(
			context.Context,
			service.CreateProductInput,
		) (domain.Product, error) {
			return domain.Product{},
				service.ErrConflict
		},
	}

	h := NewProductHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/products", h.Create)

	body := []byte(
		`{"category_id":"` +
			categoryID.String() +
			`","name":"Product","product_code":"P001"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/products",
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
			"expected status %d, got %d",
			http.StatusConflict,
			response.Code,
		)
	}
}

func TestProductHandlerCreateInvalidReference(t *testing.T) {
	gin.SetMode(gin.TestMode)

	categoryID := uuid.New()

	fake := &fakeProductService{
		createFn: func(
			context.Context,
			service.CreateProductInput,
		) (domain.Product, error) {
			return domain.Product{},
				service.ErrInvalidReference
		},
	}

	h := NewProductHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/products", h.Create)

	body := []byte(
		`{"category_id":"` +
			categoryID.String() +
			`","name":"Product","product_code":"P001"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/products",
		bytes.NewReader(body),
	)
	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnprocessableEntity,
			response.Code,
		)
	}
}

func TestProductHandlerList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fake := &fakeProductService{
		listFn: func(
			_ context.Context,
			input service.ListProductsInput,
		) (service.Page[domain.Product], error) {
			if input.Page.Limit != 10 {
				t.Fatalf("expected limit 10, got %d", input.Page.Limit)
			}

			next := "encoded-cursor"
			return service.Page[domain.Product]{
				Items: []domain.Product{
					{
						ID:          uuid.New(),
						Name:        "Shirt",
						ProductCode: "P001",
						IsActive:    true,
					},
				},
				NextCursor: &next,
			}, nil
		},
	}

	h := NewProductHandler(fake, zerolog.Nop())
	r := gin.New()
	r.GET("/products", h.List)

	req := httptest.NewRequest(http.MethodGet, "/products?limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
