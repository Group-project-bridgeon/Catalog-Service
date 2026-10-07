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

type fakeAttributeRepository struct {
	createFn func(
		context.Context,
		repository.CreateAttributeParams,
	) (domain.Attribute, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (domain.Attribute, error)

	getByCodeFn func(
		context.Context,
		string,
	) (domain.Attribute, error)

	getByNameFn func(
		context.Context,
		string,
	) (domain.Attribute, error)

	listFn func(
		context.Context,
	) ([]domain.Attribute, error)

	updateFn func(
		context.Context,
		repository.UpdateAttributeParams,
	) (domain.Attribute, error)
}

func (f *fakeAttributeRepository) Create(
	ctx context.Context,
	params repository.CreateAttributeParams,
) (domain.Attribute, error) {
	if f.createFn == nil {
		return domain.Attribute{}, nil
	}

	return f.createFn(ctx, params)
}

func (f *fakeAttributeRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Attribute, error) {
	if f.getByIDFn == nil {
		return domain.Attribute{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeAttributeRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Attribute, error) {
	if f.getByCodeFn == nil {
		return domain.Attribute{}, nil
	}

	return f.getByCodeFn(ctx, code)
}

func (f *fakeAttributeRepository) GetByName(
	ctx context.Context,
	name string,
) (domain.Attribute, error) {
	if f.getByNameFn == nil {
		return domain.Attribute{}, nil
	}

	return f.getByNameFn(ctx, name)
}

func (f *fakeAttributeRepository) List(
	ctx context.Context,
) ([]domain.Attribute, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakeAttributeRepository) Update(
	ctx context.Context,
	params repository.UpdateAttributeParams,
) (domain.Attribute, error) {
	if f.updateFn == nil {
		return domain.Attribute{}, nil
	}

	return f.updateFn(ctx, params)
}

func TestAttributeServiceCreate(
	t *testing.T,
) {
	repo := &fakeAttributeRepository{
		createFn: func(
			_ context.Context,
			params repository.CreateAttributeParams,
		) (domain.Attribute, error) {
			if params.ID == uuid.Nil {
				t.Fatal("expected generated ID")
			}

			if params.Name != "Color" {
				t.Fatalf(
					"unexpected name: %q",
					params.Name,
				)
			}

			if params.Code != "COLOR" {
				t.Fatalf(
					"unexpected code: %q",
					params.Code,
				)
			}

			return domain.Attribute{
				ID:   params.ID,
				Name: params.Name,
				Code: params.Code,
			}, nil
		},
	}

	s := NewAttributeService(repo)

	attribute, err := s.Create(
		context.Background(),
		CreateAttributeInput{
			Name: " Color ",
			Code: " COLOR ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attribute.Name != "Color" {
		t.Fatalf(
			"unexpected name: %s",
			attribute.Name,
		)
	}
}

func TestAttributeServiceCreateValidation(
	t *testing.T,
) {
	tests := []CreateAttributeInput{
		{
			Name: "",
			Code: "COLOR",
		},
		{
			Name: strings.Repeat("a", 101),
			Code: "COLOR",
		},
		{
			Name: "Color",
			Code: "",
		},
		{
			Name: "Color",
			Code: strings.Repeat("a", 51),
		},
	}

	s := NewAttributeService(
		&fakeAttributeRepository{},
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

func TestAttributeServiceCreateConflict(
	t *testing.T,
) {
	repo := &fakeAttributeRepository{
		createFn: func(
			context.Context,
			repository.CreateAttributeParams,
		) (domain.Attribute, error) {
			return domain.Attribute{},
				repository.ErrConflict
		},
	}

	s := NewAttributeService(repo)

	_, err := s.Create(
		context.Background(),
		CreateAttributeInput{
			Name: "Color",
			Code: "COLOR",
		},
	)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf(
			"expected ErrConflict, got %v",
			err,
		)
	}
}

func TestAttributeServiceGetByIDNotFound(
	t *testing.T,
) {
	repo := &fakeAttributeRepository{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.Attribute, error) {
			return domain.Attribute{},
				repository.ErrNotFound
		},
	}

	s := NewAttributeService(repo)

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

func TestAttributeServiceGetByCode(
	t *testing.T,
) {
	repo := &fakeAttributeRepository{
		getByCodeFn: func(
			_ context.Context,
			code string,
		) (domain.Attribute, error) {
			if code != "COLOR" {
				t.Fatalf(
					"unexpected code: %q",
					code,
				)
			}

			return domain.Attribute{
				Code: code,
			}, nil
		},
	}

	s := NewAttributeService(repo)

	_, err := s.GetByCode(
		context.Background(),
		" COLOR ",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAttributeServiceGetByName(
	t *testing.T,
) {
	repo := &fakeAttributeRepository{
		getByNameFn: func(
			_ context.Context,
			name string,
		) (domain.Attribute, error) {
			if name != "Color" {
				t.Fatalf(
					"unexpected name: %q",
					name,
				)
			}

			return domain.Attribute{
				Name: name,
			}, nil
		},
	}

	s := NewAttributeService(repo)

	_, err := s.GetByName(
		context.Background(),
		" Color ",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAttributeServiceList(
	t *testing.T,
) {
	repo := &fakeAttributeRepository{
		listFn: func(
			context.Context,
		) ([]domain.Attribute, error) {
			return []domain.Attribute{
				{
					ID:   uuid.New(),
					Name: "Color",
					Code: "COLOR",
				},
				{
					ID:   uuid.New(),
					Name: "Size",
					Code: "SIZE",
				},
			}, nil
		},
	}

	s := NewAttributeService(repo)

	attributes, err := s.List(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(attributes) != 2 {
		t.Fatalf(
			"expected 2 attributes, got %d",
			len(attributes),
		)
	}
}

func TestAttributeServiceUpdate(
	t *testing.T,
) {
	id := uuid.New()

	repo := &fakeAttributeRepository{
		updateFn: func(
			_ context.Context,
			params repository.UpdateAttributeParams,
		) (domain.Attribute, error) {
			if params.ID != id {
				t.Fatal("unexpected ID")
			}

			if params.Name != "Updated Color" {
				t.Fatalf(
					"unexpected name: %q",
					params.Name,
				)
			}

			return domain.Attribute{
				ID:   params.ID,
				Name: params.Name,
				Code: params.Code,
			}, nil
		},
	}

	s := NewAttributeService(repo)

	attribute, err := s.Update(
		context.Background(),
		id,
		UpdateAttributeInput{
			Name: " Updated Color ",
			Code: " COLOR-UPDATED ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attribute.Code != "COLOR-UPDATED" {
		t.Fatalf(
			"unexpected code: %s",
			attribute.Code,
		)
	}
}
