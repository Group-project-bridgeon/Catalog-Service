package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type fakeProductImageRepository struct {
	createFn func(
		context.Context,
		repository.CreateProductImageParams,
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
		repository.UpdateProductImageParams,
	) (domain.ProductImage, error)

	variantBelongsFn func(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (bool, error)
}

func (f *fakeProductImageRepository) Create(
	ctx context.Context,
	params repository.CreateProductImageParams,
) (domain.ProductImage, error) {
	if f.createFn == nil {
		return domain.ProductImage{}, nil
	}

	return f.createFn(ctx, params)
}

func (f *fakeProductImageRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.ProductImage, error) {
	if f.getByIDFn == nil {
		return domain.ProductImage{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeProductImageRepository) List(
	ctx context.Context,
) ([]domain.ProductImage, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakeProductImageRepository) ListByProductID(
	ctx context.Context,
	id uuid.UUID,
) ([]domain.ProductImage, error) {
	if f.listByProductIDFn == nil {
		return nil, nil
	}

	return f.listByProductIDFn(ctx, id)
}

func (f *fakeProductImageRepository) ListByVariantID(
	ctx context.Context,
	id uuid.UUID,
) ([]domain.ProductImage, error) {
	if f.listByVariantIDFn == nil {
		return nil, nil
	}

	return f.listByVariantIDFn(ctx, id)
}

func (f *fakeProductImageRepository) Update(
	ctx context.Context,
	params repository.UpdateProductImageParams,
) (domain.ProductImage, error) {
	if f.updateFn == nil {
		return domain.ProductImage{}, nil
	}

	return f.updateFn(ctx, params)
}

func (f *fakeProductImageRepository) VariantBelongsToProduct(
	ctx context.Context,
	productID uuid.UUID,
	variantID uuid.UUID,
) (bool, error) {
	if f.variantBelongsFn == nil {
		return true, nil
	}

	return f.variantBelongsFn(
		ctx,
		productID,
		variantID,
	)
}

func TestProductImageServiceCreate(
	t *testing.T,
) {
	productID := uuid.New()
	variantID := uuid.New()

	repo := &fakeProductImageRepository{
		createFn: func(
			_ context.Context,
			params repository.CreateProductImageParams,
		) (domain.ProductImage, error) {
			if params.ID == uuid.Nil {
				t.Fatal("expected generated ID")
			}

			if params.ProductID != productID {
				t.Fatal("unexpected product ID")
			}

			if params.VariantID == nil ||
				*params.VariantID != variantID {
				t.Fatal("unexpected variant ID")
			}

			if params.ImageURL !=
				"https://example.com/image.jpg" {
				t.Fatal("unexpected image URL")
			}

			if params.SortOrder != 2 {
				t.Fatal("unexpected sort order")
			}

			return domain.ProductImage{
				ID:        params.ID,
				ProductID: params.ProductID,
				VariantID: params.VariantID,
				ImageURL:  params.ImageURL,
				SortOrder: params.SortOrder,
			}, nil
		},
	}

	s := NewProductImageService(repo)

	_, err := s.Create(
		context.Background(),
		CreateProductImageInput{
			ProductID: productID,
			VariantID: &variantID,
			ImageURL:  " https://example.com/image.jpg ",
			SortOrder: 2,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductImageServiceValidation(
	t *testing.T,
) {
	productID := uuid.New()

	tests := []CreateProductImageInput{
		{
			ImageURL: "https://example.com/image.jpg",
		},
		{
			ProductID: productID,
			ImageURL:  "",
		},
		{
			ProductID: productID,
			ImageURL:  "not-a-url",
		},
		{
			ProductID: productID,
			ImageURL:  "ftp://example.com/image.jpg",
		},
	}

	s := NewProductImageService(
		&fakeProductImageRepository{},
	)

	for index, input := range tests {
		_, err := s.Create(
			context.Background(),
			input,
		)

		if !errors.Is(
			err,
			ErrInvalidArgument,
		) {
			t.Fatalf(
				"case %d: expected ErrInvalidArgument, got %v",
				index,
				err,
			)
		}
	}
}

func TestProductImageServiceVariantMustBelongToProduct(
	t *testing.T,
) {
	variantID := uuid.New()

	repo := &fakeProductImageRepository{
		variantBelongsFn: func(
			context.Context,
			uuid.UUID,
			uuid.UUID,
		) (bool, error) {
			return false, nil
		},
	}

	s := NewProductImageService(repo)

	_, err := s.Create(
		context.Background(),
		CreateProductImageInput{
			ProductID: uuid.New(),
			VariantID: &variantID,
			ImageURL:  "https://example.com/image.jpg",
		},
	)

	if !errors.Is(
		err,
		ErrInvalidReference,
	) {
		t.Fatalf(
			"expected ErrInvalidReference, got %v",
			err,
		)
	}
}

func TestProductImageServiceGetNotFound(
	t *testing.T,
) {
	repo := &fakeProductImageRepository{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.ProductImage, error) {
			return domain.ProductImage{},
				repository.ErrNotFound
		},
	}

	s := NewProductImageService(repo)

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

func TestProductImageServiceUpdate(
	t *testing.T,
) {
	id := uuid.New()
	productID := uuid.New()

	repo := &fakeProductImageRepository{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.ProductImage, error) {
			return domain.ProductImage{
				ID:        id,
				ProductID: productID,
			}, nil
		},

		updateFn: func(
			_ context.Context,
			params repository.UpdateProductImageParams,
		) (domain.ProductImage, error) {
			if params.ImageURL !=
				"https://example.com/updated.jpg" {
				t.Fatal("unexpected URL")
			}

			if params.SortOrder != 5 {
				t.Fatal("unexpected sort order")
			}

			return domain.ProductImage{
				ID:        params.ID,
				ProductID: productID,
				VariantID: params.VariantID,
				ImageURL:  params.ImageURL,
				SortOrder: params.SortOrder,
			}, nil
		},
	}

	s := NewProductImageService(repo)

	image, err := s.Update(
		context.Background(),
		id,
		UpdateProductImageInput{
			ImageURL:  "https://example.com/updated.jpg",
			SortOrder: 5,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if image.SortOrder != 5 {
		t.Fatal("unexpected sort order")
	}
}
