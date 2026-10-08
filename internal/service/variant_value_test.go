package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type fakeVariantValueRepository struct {
	createFn func(
		context.Context,
		repository.CreateVariantValueParams,
	) (domain.VariantValue, error)

	getFn func(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (domain.VariantValue, error)

	listFn func(
		context.Context,
		repository.ListVariantValuesParams,
	) ([]domain.VariantValue, error)

	updateFn func(
		context.Context,
		repository.UpdateVariantValueParams,
	) (domain.VariantValue, error)

	combinationExistsFn func(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		uuid.UUID,
	) (bool, error)
}

func (f *fakeVariantValueRepository) Create(
	ctx context.Context,
	params repository.CreateVariantValueParams,
) (domain.VariantValue, error) {
	if f.createFn == nil {
		return domain.VariantValue{}, nil
	}

	return f.createFn(ctx, params)
}

func (f *fakeVariantValueRepository) GetByVariantAndAttribute(
	ctx context.Context,
	variantID uuid.UUID,
	attributeID uuid.UUID,
) (domain.VariantValue, error) {
	if f.getFn == nil {
		return domain.VariantValue{}, nil
	}

	return f.getFn(
		ctx,
		variantID,
		attributeID,
	)
}

func (f *fakeVariantValueRepository) List(
	ctx context.Context,
	params repository.ListVariantValuesParams,
) ([]domain.VariantValue, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx, params)
}

func (f *fakeVariantValueRepository) Update(
	ctx context.Context,
	params repository.UpdateVariantValueParams,
) (domain.VariantValue, error) {
	if f.updateFn == nil {
		return domain.VariantValue{}, nil
	}

	return f.updateFn(ctx, params)
}

func (f *fakeVariantValueRepository) CombinationExists(
	ctx context.Context,
	variantID uuid.UUID,
	attributeID uuid.UUID,
	attributeValueID uuid.UUID,
) (bool, error) {
	if f.combinationExistsFn == nil {
		return false, nil
	}

	return f.combinationExistsFn(
		ctx,
		variantID,
		attributeID,
		attributeValueID,
	)
}

func TestVariantValueServiceCreate(
	t *testing.T,
) {
	variantID := uuid.New()
	attributeID := uuid.New()
	valueID := uuid.New()

	repo := &fakeVariantValueRepository{
		createFn: func(
			_ context.Context,
			params repository.CreateVariantValueParams,
		) (domain.VariantValue, error) {
			if params.VariantID != variantID {
				t.Fatal("unexpected variant ID")
			}

			if params.AttributeID != attributeID {
				t.Fatal("unexpected attribute ID")
			}

			if params.AttributeValueID != valueID {
				t.Fatal("unexpected value ID")
			}

			return domain.VariantValue{
				VariantID:        params.VariantID,
				AttributeID:      params.AttributeID,
				AttributeValueID: params.AttributeValueID,
			}, nil
		},
	}

	s := NewVariantValueService(repo)

	result, err := s.Create(
		context.Background(),
		CreateVariantValueInput{
			VariantID:        variantID,
			AttributeID:      attributeID,
			AttributeValueID: valueID,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AttributeValueID != valueID {
		t.Fatal("unexpected attribute value ID")
	}
}

func TestVariantValueServiceValidation(
	t *testing.T,
) {
	valid := uuid.New()

	tests := []CreateVariantValueInput{
		{
			AttributeID:      valid,
			AttributeValueID: valid,
		},
		{
			VariantID:        valid,
			AttributeValueID: valid,
		},
		{
			VariantID:   valid,
			AttributeID: valid,
		},
	}

	s := NewVariantValueService(
		&fakeVariantValueRepository{},
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

func TestVariantValueServiceDuplicateCombination(
	t *testing.T,
) {
	repo := &fakeVariantValueRepository{
		combinationExistsFn: func(
			context.Context,
			uuid.UUID,
			uuid.UUID,
			uuid.UUID,
		) (bool, error) {
			return true, nil
		},
	}

	s := NewVariantValueService(repo)

	_, err := s.Create(
		context.Background(),
		CreateVariantValueInput{
			VariantID:        uuid.New(),
			AttributeID:      uuid.New(),
			AttributeValueID: uuid.New(),
		},
	)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf(
			"expected ErrConflict, got %v",
			err,
		)
	}
}

func TestVariantValueServiceInvalidReference(
	t *testing.T,
) {
	repo := &fakeVariantValueRepository{
		createFn: func(
			context.Context,
			repository.CreateVariantValueParams,
		) (domain.VariantValue, error) {
			return domain.VariantValue{},
				repository.ErrInvalidReference
		},
	}

	s := NewVariantValueService(repo)

	_, err := s.Create(
		context.Background(),
		CreateVariantValueInput{
			VariantID:        uuid.New(),
			AttributeID:      uuid.New(),
			AttributeValueID: uuid.New(),
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

func TestVariantValueServiceGetNotFound(
	t *testing.T,
) {
	repo := &fakeVariantValueRepository{
		getFn: func(
			context.Context,
			uuid.UUID,
			uuid.UUID,
		) (domain.VariantValue, error) {
			return domain.VariantValue{},
				repository.ErrNotFound
		},
	}

	s := NewVariantValueService(repo)

	_, err := s.GetByVariantAndAttribute(
		context.Background(),
		uuid.New(),
		uuid.New(),
	)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound, got %v",
			err,
		)
	}
}

func TestVariantValueServiceUpdate(
	t *testing.T,
) {
	variantID := uuid.New()
	attributeID := uuid.New()
	valueID := uuid.New()

	repo := &fakeVariantValueRepository{
		updateFn: func(
			_ context.Context,
			params repository.UpdateVariantValueParams,
		) (domain.VariantValue, error) {
			if params.AttributeValueID != valueID {
				t.Fatal("unexpected value ID")
			}

			return domain.VariantValue{
				VariantID:        params.VariantID,
				AttributeID:      params.AttributeID,
				AttributeValueID: params.AttributeValueID,
			}, nil
		},
	}

	s := NewVariantValueService(repo)

	result, err := s.Update(
		context.Background(),
		variantID,
		attributeID,
		UpdateVariantValueInput{
			AttributeValueID: valueID,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AttributeValueID != valueID {
		t.Fatal("unexpected value ID")
	}
}

func TestVariantValueServiceList(t *testing.T) {
	variantID := uuid.New()

	repo := &fakeVariantValueRepository{
		listFn: func(
			_ context.Context,
			params repository.ListVariantValuesParams,
		) ([]domain.VariantValue, error) {
			if params.PageSize != 3 {
				t.Fatalf("expected PageSize 3, got %d", params.PageSize)
			}
			if params.VariantID == nil || *params.VariantID != variantID {
				t.Fatalf("expected VariantID %v", variantID)
			}

			return []domain.VariantValue{
				{VariantID: variantID, AttributeID: uuid.New(), AttributeValueID: uuid.New()},
				{VariantID: variantID, AttributeID: uuid.New(), AttributeValueID: uuid.New()},
				{VariantID: variantID, AttributeID: uuid.New(), AttributeValueID: uuid.New()}, // extra
			}, nil
		},
	}

	s := NewVariantValueService(repo)

	page, err := s.List(context.Background(), ListVariantValuesInput{
		Page:      PageInput{Limit: 2},
		VariantID: &variantID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(page.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(page.Items))
	}
	if page.NextCursor == nil || *page.NextCursor == "" {
		t.Fatal("expected next cursor")
	}
}
