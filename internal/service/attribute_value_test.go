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

type fakeAttributeValueRepository struct {
	createFn func(
		context.Context,
		repository.CreateAttributeValueParams,
	) (domain.AttributeValue, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (domain.AttributeValue, error)

	getByAttributeAndValueFn func(
		context.Context,
		uuid.UUID,
		string,
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
		repository.UpdateAttributeValueParams,
	) (domain.AttributeValue, error)
}

func (f *fakeAttributeValueRepository) Create(
	ctx context.Context,
	params repository.CreateAttributeValueParams,
) (domain.AttributeValue, error) {
	if f.createFn == nil {
		return domain.AttributeValue{}, nil
	}

	return f.createFn(ctx, params)
}

func (f *fakeAttributeValueRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.AttributeValue, error) {
	if f.getByIDFn == nil {
		return domain.AttributeValue{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeAttributeValueRepository) GetByAttributeAndValue(
	ctx context.Context,
	attributeID uuid.UUID,
	value string,
) (domain.AttributeValue, error) {
	if f.getByAttributeAndValueFn == nil {
		return domain.AttributeValue{}, nil
	}

	return f.getByAttributeAndValueFn(
		ctx,
		attributeID,
		value,
	)
}

func (f *fakeAttributeValueRepository) List(
	ctx context.Context,
) ([]domain.AttributeValue, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakeAttributeValueRepository) ListByAttributeID(
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

func (f *fakeAttributeValueRepository) Update(
	ctx context.Context,
	params repository.UpdateAttributeValueParams,
) (domain.AttributeValue, error) {
	if f.updateFn == nil {
		return domain.AttributeValue{}, nil
	}

	return f.updateFn(ctx, params)
}

func TestAttributeValueServiceCreate(
	t *testing.T,
) {
	attributeID := uuid.New()

	repo := &fakeAttributeValueRepository{
		createFn: func(
			_ context.Context,
			params repository.CreateAttributeValueParams,
		) (domain.AttributeValue, error) {
			if params.ID == uuid.Nil {
				t.Fatal("expected generated ID")
			}

			if params.AttributeID != attributeID {
				t.Fatal("unexpected attribute ID")
			}

			if params.Value != "Blue" {
				t.Fatalf(
					"expected trimmed value, got %q",
					params.Value,
				)
			}

			return domain.AttributeValue{
				ID:          params.ID,
				AttributeID: params.AttributeID,
				Value:       params.Value,
			}, nil
		},
	}

	s := NewAttributeValueService(repo)

	result, err := s.Create(
		context.Background(),
		CreateAttributeValueInput{
			AttributeID: attributeID,
			Value:       " Blue ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Value != "Blue" {
		t.Fatalf(
			"unexpected value: %s",
			result.Value,
		)
	}
}

func TestAttributeValueServiceValidation(
	t *testing.T,
) {
	attributeID := uuid.New()

	tests := []CreateAttributeValueInput{
		{
			Value: "Blue",
		},
		{
			AttributeID: attributeID,
			Value:       " ",
		},
		{
			AttributeID: attributeID,
			Value: strings.Repeat(
				"a",
				101,
			),
		},
	}

	s := NewAttributeValueService(
		&fakeAttributeValueRepository{},
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

func TestAttributeValueServiceConflict(
	t *testing.T,
) {
	repo := &fakeAttributeValueRepository{
		createFn: func(
			context.Context,
			repository.CreateAttributeValueParams,
		) (domain.AttributeValue, error) {
			return domain.AttributeValue{},
				repository.ErrConflict
		},
	}

	s := NewAttributeValueService(repo)

	_, err := s.Create(
		context.Background(),
		CreateAttributeValueInput{
			AttributeID: uuid.New(),
			Value:       "Blue",
		},
	)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf(
			"expected ErrConflict, got %v",
			err,
		)
	}
}

func TestAttributeValueServiceInvalidReference(
	t *testing.T,
) {
	repo := &fakeAttributeValueRepository{
		createFn: func(
			context.Context,
			repository.CreateAttributeValueParams,
		) (domain.AttributeValue, error) {
			return domain.AttributeValue{},
				repository.ErrInvalidReference
		},
	}

	s := NewAttributeValueService(repo)

	_, err := s.Create(
		context.Background(),
		CreateAttributeValueInput{
			AttributeID: uuid.New(),
			Value:       "Blue",
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

func TestAttributeValueServiceGetByIDNotFound(
	t *testing.T,
) {
	repo := &fakeAttributeValueRepository{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.AttributeValue, error) {
			return domain.AttributeValue{},
				repository.ErrNotFound
		},
	}

	s := NewAttributeValueService(repo)

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

func TestAttributeValueServiceListByAttributeID(
	t *testing.T,
) {
	attributeID := uuid.New()

	repo := &fakeAttributeValueRepository{
		listByAttributeIDFn: func(
			_ context.Context,
			gotID uuid.UUID,
		) ([]domain.AttributeValue, error) {
			if gotID != attributeID {
				t.Fatal("unexpected attribute ID")
			}

			return []domain.AttributeValue{
				{
					ID:          uuid.New(),
					AttributeID: attributeID,
					Value:       "Blue",
				},
				{
					ID:          uuid.New(),
					AttributeID: attributeID,
					Value:       "Red",
				},
			}, nil
		},
	}

	s := NewAttributeValueService(repo)

	values, err := s.ListByAttributeID(
		context.Background(),
		attributeID,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(values) != 2 {
		t.Fatalf(
			"expected 2 values, got %d",
			len(values),
		)
	}
}

func TestAttributeValueServiceUpdate(
	t *testing.T,
) {
	id := uuid.New()

	repo := &fakeAttributeValueRepository{
		updateFn: func(
			_ context.Context,
			params repository.UpdateAttributeValueParams,
		) (domain.AttributeValue, error) {
			if params.ID != id {
				t.Fatal("unexpected ID")
			}

			if params.Value != "Navy Blue" {
				t.Fatalf(
					"unexpected value: %q",
					params.Value,
				)
			}

			return domain.AttributeValue{
				ID:    params.ID,
				Value: params.Value,
			}, nil
		},
	}

	s := NewAttributeValueService(repo)

	result, err := s.Update(
		context.Background(),
		id,
		UpdateAttributeValueInput{
			Value: " Navy Blue ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Value != "Navy Blue" {
		t.Fatalf(
			"unexpected value: %s",
			result.Value,
		)
	}
}
