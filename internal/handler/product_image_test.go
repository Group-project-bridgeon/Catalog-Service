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

type fakeProductImageService struct {
	createFn func(
		context.Context,
		service.CreateProductImageInput,
	) (domain.ProductImage, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (domain.ProductImage, error)

	listFn func(
		context.Context,
	) ([]domain.ProductImage, error)

	listByProductIDFn func(
		context.Context,
		uuid.UUID,
	) ([]domain.ProductImage, error)

	listByVariantIDFn func(
		context.Context,
		uuid.UUID,
	) ([]domain.ProductImage, error)

	updateFn func(
		context.Context,
		uuid.UUID,
		service.UpdateProductImageInput,
	) (domain.ProductImage, error)
}

func (f *fakeProductImageService) Create(
	ctx context.Context,
	input service.CreateProductImageInput,
) (domain.ProductImage, error) {
	if f.createFn == nil {
		return domain.ProductImage{}, nil
	}

	return f.createFn(ctx, input)
}

func (f *fakeProductImageService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.ProductImage, error) {
	if f.getByIDFn == nil {
		return domain.ProductImage{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeProductImageService) List(
	ctx context.Context,
) ([]domain.ProductImage, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakeProductImageService) ListByProductID(
	ctx context.Context,
	id uuid.UUID,
) ([]domain.ProductImage, error) {
	if f.listByProductIDFn == nil {
		return nil, nil
	}

	return f.listByProductIDFn(ctx, id)
}

func (f *fakeProductImageService) ListByVariantID(
	ctx context.Context,
	id uuid.UUID,
) ([]domain.ProductImage, error) {
	if f.listByVariantIDFn == nil {
		return nil, nil
	}

	return f.listByVariantIDFn(ctx, id)
}

func (f *fakeProductImageService) Update(
	ctx context.Context,
	id uuid.UUID,
	input service.UpdateProductImageInput,
) (domain.ProductImage, error) {
	if f.updateFn == nil {
		return domain.ProductImage{}, nil
	}

	return f.updateFn(ctx, id, input)
}

func TestProductImageHandlerCreate(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	productID := uuid.New()
	id := uuid.New()

	fake := &fakeProductImageService{
		createFn: func(
			_ context.Context,
			input service.CreateProductImageInput,
		) (domain.ProductImage, error) {
			return domain.ProductImage{
				ID:        id,
				ProductID: input.ProductID,
				VariantID: input.VariantID,
				ImageURL:  input.ImageURL,
				SortOrder: input.SortOrder,
			}, nil
		},
	}

	h := NewProductImageHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/product-images", h.Create)

	body := []byte(
		`{"product_id":"` +
			productID.String() +
			`","image_url":"https://example.com/image.jpg","sort_order":1}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/product-images",
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

func TestProductImageHandlerInvalidUUID(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	h := NewProductImageHandler(
		&fakeProductImageService{},
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/product-images/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/product-images/not-a-uuid",
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

func TestProductImageHandlerNotFound(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakeProductImageService{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.ProductImage, error) {
			return domain.ProductImage{},
				service.ErrNotFound
		},
	}

	h := NewProductImageHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.GET("/product-images/:id", h.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/product-images/"+uuid.New().String(),
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

func TestProductImageHandlerInvalidReference(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	fake := &fakeProductImageService{
		createFn: func(
			context.Context,
			service.CreateProductImageInput,
		) (domain.ProductImage, error) {
			return domain.ProductImage{},
				service.ErrInvalidReference
		},
	}

	h := NewProductImageHandler(
		fake,
		zerolog.Nop(),
	)

	r := gin.New()
	r.POST("/product-images", h.Create)

	body := []byte(
		`{"product_id":"` +
			uuid.New().String() +
			`","variant_id":"` +
			uuid.New().String() +
			`","image_url":"https://example.com/image.jpg"}`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/product-images",
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
