package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type fakeProductRepository struct {
	createFn    func(context.Context, repository.CreateProductParams) (domain.Product, error)
	getByIDFn   func(context.Context, uuid.UUID) (domain.Product, error)
	getByCodeFn func(context.Context, string) (domain.Product, error)
	listFn      func(context.Context, repository.ListProductsParams) ([]domain.Product, error)
	updateFn    func(context.Context, repository.UpdateProductParams) (domain.Product, error)
	setActiveFn func(context.Context, uuid.UUID, bool) (domain.Product, error)
}

func (f *fakeProductRepository) Create(
	ctx context.Context,
	params repository.CreateProductParams,
) (domain.Product, error) {
	if f.createFn == nil {
		return domain.Product{}, nil
	}

	return f.createFn(ctx, params)
}

func (f *fakeProductRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Product, error) {
	if f.getByIDFn == nil {
		return domain.Product{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeProductRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Product, error) {
	if f.getByCodeFn == nil {
		return domain.Product{}, nil
	}

	return f.getByCodeFn(ctx, code)
}

func (f *fakeProductRepository) List(
	ctx context.Context,
	params repository.ListProductsParams,
) ([]domain.Product, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx, params)
}

func (f *fakeProductRepository) Update(
	ctx context.Context,
	params repository.UpdateProductParams,
) (domain.Product, error) {
	if f.updateFn == nil {
		return domain.Product{}, nil
	}

	return f.updateFn(ctx, params)
}

func (f *fakeProductRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Product, error) {
	if f.setActiveFn == nil {
		return domain.Product{}, nil
	}

	return f.setActiveFn(ctx, id, isActive)
}

func TestProductServiceCreate(t *testing.T) {
	categoryID := uuid.New()
	brandID := uuid.New()

	description := "  Cotton shirt  "

	repo := &fakeProductRepository{
		createFn: func(
			_ context.Context,
			params repository.CreateProductParams,
		) (domain.Product, error) {
			if params.ID == uuid.Nil {
				t.Fatal("expected generated product ID")
			}

			if params.CategoryID != categoryID {
				t.Fatal("unexpected category ID")
			}

			if params.BrandID == nil || *params.BrandID != brandID {
				t.Fatal("unexpected brand ID")
			}

			if params.Name != "Formal Shirt" {
				t.Fatalf(
					"expected trimmed name, got %q",
					params.Name,
				)
			}

			if params.ProductCode != "SHIRT-001" {
				t.Fatalf(
					"expected trimmed code, got %q",
					params.ProductCode,
				)
			}

			if params.Description == nil ||
				*params.Description != "Cotton shirt" {
				t.Fatal("expected trimmed description")
			}

			if !params.IsActive {
				t.Fatal("new product should be active")
			}

			return domain.Product{
				ID:          params.ID,
				CategoryID:  params.CategoryID,
				BrandID:     params.BrandID,
				Name:        params.Name,
				ProductCode: params.ProductCode,
				Description: params.Description,
				IsActive:    params.IsActive,
			}, nil
		},
	}

	s := NewProductService(repo)

	product, err := s.Create(
		context.Background(),
		CreateProductInput{
			CategoryID:  categoryID,
			BrandID:     &brandID,
			Name:        "  Formal Shirt  ",
			ProductCode: "  SHIRT-001  ",
			Description: &description,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if product.ProductCode != "SHIRT-001" {
		t.Fatalf(
			"unexpected product code: %s",
			product.ProductCode,
		)
	}
}

func TestProductServiceCreateValidation(t *testing.T) {
	categoryID := uuid.New()
	nilBrandID := uuid.Nil

	tests := []struct {
		name  string
		input CreateProductInput
	}{
		{
			name: "missing category",
			input: CreateProductInput{
				Name:        "Product",
				ProductCode: "P001",
			},
		},
		{
			name: "invalid brand id",
			input: CreateProductInput{
				CategoryID:  categoryID,
				BrandID:     &nilBrandID,
				Name:        "Product",
				ProductCode: "P001",
			},
		},
		{
			name: "empty name",
			input: CreateProductInput{
				CategoryID:  categoryID,
				Name:        " ",
				ProductCode: "P001",
			},
		},
		{
			name: "name too long",
			input: CreateProductInput{
				CategoryID:  categoryID,
				Name:        strings.Repeat("a", 201),
				ProductCode: "P001",
			},
		},
		{
			name: "empty code",
			input: CreateProductInput{
				CategoryID:  categoryID,
				Name:        "Product",
				ProductCode: " ",
			},
		},
		{
			name: "code too long",
			input: CreateProductInput{
				CategoryID:  categoryID,
				Name:        "Product",
				ProductCode: strings.Repeat("a", 51),
			},
		},
	}

	s := NewProductService(&fakeProductRepository{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.Create(
				context.Background(),
				tt.input,
			)

			if !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf(
					"expected ErrInvalidArgument, got %v",
					err,
				)
			}
		})
	}
}

func TestProductServiceCreateConflict(t *testing.T) {
	repo := &fakeProductRepository{
		createFn: func(
			context.Context,
			repository.CreateProductParams,
		) (domain.Product, error) {
			return domain.Product{}, repository.ErrConflict
		},
	}

	s := NewProductService(repo)

	_, err := s.Create(
		context.Background(),
		CreateProductInput{
			CategoryID:  uuid.New(),
			Name:        "Product",
			ProductCode: "P001",
		},
	)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf(
			"expected ErrConflict, got %v",
			err,
		)
	}
}

func TestProductServiceCreateInvalidReference(t *testing.T) {
	repo := &fakeProductRepository{
		createFn: func(
			context.Context,
			repository.CreateProductParams,
		) (domain.Product, error) {
			return domain.Product{},
				repository.ErrInvalidReference
		},
	}

	s := NewProductService(repo)

	_, err := s.Create(
		context.Background(),
		CreateProductInput{
			CategoryID:  uuid.New(),
			Name:        "Product",
			ProductCode: "P001",
		},
	)

	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf(
			"expected ErrInvalidReference, got %v",
			err,
		)
	}
}

func TestProductServiceGetByIDNotFound(t *testing.T) {
	repo := &fakeProductRepository{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.Product, error) {
			return domain.Product{},
				repository.ErrNotFound
		},
	}

	s := NewProductService(repo)

	_, err := s.GetByID(
		context.Background(),
		uuid.New(),
	)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound, got %v",
			err,
		)
	}
}

func TestProductServiceGetByCode(t *testing.T) {
	repo := &fakeProductRepository{
		getByCodeFn: func(
			_ context.Context,
			code string,
		) (domain.Product, error) {
			if code != "P001" {
				t.Fatalf(
					"expected trimmed code, got %q",
					code,
				)
			}

			return domain.Product{
				ProductCode: code,
			}, nil
		},
	}

	s := NewProductService(repo)

	_, err := s.GetByCode(
		context.Background(),
		"  P001  ",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductServiceList(t *testing.T) {
	categoryID := uuid.New()
	brandID := uuid.New()
	isActive := true

	repo := &fakeProductRepository{
		listFn: func(
			_ context.Context,
			params repository.ListProductsParams,
		) ([]domain.Product, error) {
			if params.PageSize != 3 { // limit (2) + 1
				t.Fatalf("expected PageSize 3, got %d", params.PageSize)
			}
			if params.CategoryID == nil || *params.CategoryID != categoryID {
				t.Fatalf("expected CategoryID %v", categoryID)
			}
			if params.BrandID == nil || *params.BrandID != brandID {
				t.Fatalf("expected BrandID %v", brandID)
			}
			if params.IsActive == nil || *params.IsActive != isActive {
				t.Fatalf("expected IsActive %v", isActive)
			}

			return []domain.Product{
				{ID: uuid.New(), Name: "Product A"},
				{ID: uuid.New(), Name: "Product B"},
				{ID: uuid.New(), Name: "Product C"}, // Extra item -> hasMore = true
			}, nil
		},
	}

	s := NewProductService(repo)

	page, err := s.List(context.Background(), ListProductsInput{
		Page: PageInput{
			Limit: 2,
		},
		CategoryID: &categoryID,
		BrandID:    &brandID,
		IsActive:   &isActive,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(page.Items) != 2 {
		t.Fatalf("expected 2 products after trim, got %d", len(page.Items))
	}

	if page.NextCursor == nil || *page.NextCursor == "" {
		t.Fatal("expected next cursor to be generated")
	}

	// Test invalid cursor
	_, err = s.List(context.Background(), ListProductsInput{
		Page: PageInput{
			Limit:  2,
			Cursor: "invalid-base64",
		},
	})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for invalid cursor, got %v", err)
	}

	// Test limit validation (0 -> default limit 50, >100 -> error)
	_, err = s.List(context.Background(), ListProductsInput{
		Page: PageInput{
			Limit: 101,
		},
	})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for limit > 100, got %v", err)
	}
}

func TestProductServiceUpdate(t *testing.T) {
	id := uuid.New()
	categoryID := uuid.New()

	repo := &fakeProductRepository{
		updateFn: func(
			_ context.Context,
			params repository.UpdateProductParams,
		) (domain.Product, error) {
			if params.ID != id {
				t.Fatal("unexpected product ID")
			}

			if params.Name != "Updated Product" {
				t.Fatalf(
					"unexpected name: %q",
					params.Name,
				)
			}

			return domain.Product{
				ID:          params.ID,
				CategoryID:  params.CategoryID,
				Name:        params.Name,
				ProductCode: params.ProductCode,
			}, nil
		},
	}

	s := NewProductService(repo)

	_, err := s.Update(
		context.Background(),
		id,
		UpdateProductInput{
			CategoryID:  categoryID,
			Name:        " Updated Product ",
			ProductCode: " P002 ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductServiceSetActive(t *testing.T) {
	id := uuid.New()

	repo := &fakeProductRepository{
		setActiveFn: func(
			_ context.Context,
			gotID uuid.UUID,
			active bool,
		) (domain.Product, error) {
			if gotID != id {
				t.Fatal("unexpected product ID")
			}

			if active {
				t.Fatal("expected false")
			}

			return domain.Product{
				ID:       gotID,
				IsActive: active,
			}, nil
		},
	}

	s := NewProductService(repo)

	product, err := s.SetActive(
		context.Background(),
		id,
		false,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if product.IsActive {
		t.Fatal("expected inactive product")
	}
}
