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

type fakeCategoryRepository struct {
	createFn    func(context.Context, repository.CreateCategoryParams) (domain.Category, error)
	getByIDFn   func(context.Context, uuid.UUID) (domain.Category, error)
	getByCodeFn func(context.Context, string) (domain.Category, error)
	listFn      func(context.Context) ([]domain.Category, error)
	updateFn    func(context.Context, repository.UpdateCategoryParams) (domain.Category, error)
	setActiveFn func(context.Context, uuid.UUID, bool) (domain.Category, error)
}

func (f *fakeCategoryRepository) Create(
	ctx context.Context,
	params repository.CreateCategoryParams,
) (domain.Category, error) {
	return f.createFn(ctx, params)
}

func (f *fakeCategoryRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Category, error) {
	return f.getByIDFn(ctx, id)
}

func (f *fakeCategoryRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Category, error) {
	return f.getByCodeFn(ctx, code)
}

func (f *fakeCategoryRepository) List(
	ctx context.Context,
) ([]domain.Category, error) {
	return f.listFn(ctx)
}

func (f *fakeCategoryRepository) Update(
	ctx context.Context,
	params repository.UpdateCategoryParams,
) (domain.Category, error) {
	return f.updateFn(ctx, params)
}

func (f *fakeCategoryRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Category, error) {
	return f.setActiveFn(ctx, id, isActive)
}

func TestCategoryServiceCreate(t *testing.T) {
	var captured repository.CreateCategoryParams

	repo := &fakeCategoryRepository{
		createFn: func(
			_ context.Context,
			params repository.CreateCategoryParams,
		) (domain.Category, error) {
			captured = params

			return domain.Category{
				ID:       params.ID,
				Name:     params.Name,
				Code:     params.Code,
				IsActive: params.IsActive,
			}, nil
		},
	}

	service := NewCategoryService(repo)

	category, err := service.Create(
		context.Background(),
		CreateCategoryInput{
			Name: "  Shirts  ",
			Code: "  SHIRTS  ",
		},
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if captured.ID == uuid.Nil {
		t.Fatal("Create() did not generate category UUID")
	}

	if captured.Name != "Shirts" {
		t.Fatalf("Name = %q, want %q", captured.Name, "Shirts")
	}

	if captured.Code != "SHIRTS" {
		t.Fatalf("Code = %q, want %q", captured.Code, "SHIRTS")
	}

	if !captured.IsActive {
		t.Fatal("new category should be active")
	}

	if category.ID != captured.ID {
		t.Fatal("returned category ID does not match repository result")
	}
}

func TestCategoryServiceCreateValidation(t *testing.T) {
	tests := []struct {
		name  string
		input CreateCategoryInput
	}{
		{
			name: "empty name",
			input: CreateCategoryInput{
				Name: "",
				Code: "CODE",
			},
		},
		{
			name: "name too long",
			input: CreateCategoryInput{
				Name: strings.Repeat("a", 101),
				Code: "CODE",
			},
		},
		{
			name: "empty code",
			input: CreateCategoryInput{
				Name: "Category",
				Code: "",
			},
		},
		{
			name: "code too long",
			input: CreateCategoryInput{
				Name: "Category",
				Code: strings.Repeat("a", 51),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCategoryService(&fakeCategoryRepository{})

			_, err := service.Create(
				context.Background(),
				tt.input,
			)

			if !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf(
					"Create() error = %v, want ErrInvalidArgument",
					err,
				)
			}
		})
	}
}

func TestCategoryServiceCreateConflict(t *testing.T) {
	repo := &fakeCategoryRepository{
		createFn: func(
			context.Context,
			repository.CreateCategoryParams,
		) (domain.Category, error) {
			return domain.Category{}, repository.ErrConflict
		},
	}

	service := NewCategoryService(repo)

	_, err := service.Create(
		context.Background(),
		CreateCategoryInput{
			Name: "Shirts",
			Code: "SHIRTS",
		},
	)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf("Create() error = %v, want ErrConflict", err)
	}
}

func TestCategoryServiceGetByIDNotFound(t *testing.T) {
	repo := &fakeCategoryRepository{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.Category, error) {
			return domain.Category{}, repository.ErrNotFound
		},
	}

	service := NewCategoryService(repo)

	_, err := service.GetByID(
		context.Background(),
		uuid.New(),
	)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID() error = %v, want ErrNotFound", err)
	}
}

func TestCategoryServiceGetByCode(t *testing.T) {
	repo := &fakeCategoryRepository{
		getByCodeFn: func(
			_ context.Context,
			code string,
		) (domain.Category, error) {
			if code != "SHIRTS" {
				t.Fatalf("code = %q, want SHIRTS", code)
			}

			return domain.Category{
				Name: "Shirts",
				Code: code,
			}, nil
		},
	}

	service := NewCategoryService(repo)

	category, err := service.GetByCode(
		context.Background(),
		"  SHIRTS  ",
	)
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}

	if category.Code != "SHIRTS" {
		t.Fatalf(
			"category.Code = %q, want SHIRTS",
			category.Code,
		)
	}
}

func TestCategoryServiceList(t *testing.T) {
	repo := &fakeCategoryRepository{
		listFn: func(
			context.Context,
		) ([]domain.Category, error) {
			return []domain.Category{
				{Name: "Shirts", Code: "SHIRTS"},
				{Name: "Trousers", Code: "TROUSERS"},
			}, nil
		},
	}

	service := NewCategoryService(repo)

	categories, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(categories) != 2 {
		t.Fatalf(
			"List() length = %d, want 2",
			len(categories),
		)
	}
}

func TestCategoryServiceUpdate(t *testing.T) {
	id := uuid.New()

	repo := &fakeCategoryRepository{
		updateFn: func(
			_ context.Context,
			params repository.UpdateCategoryParams,
		) (domain.Category, error) {
			if params.ID != id {
				t.Fatal("incorrect category ID")
			}

			if params.Name != "Formal Shirts" {
				t.Fatalf(
					"Name = %q, want Formal Shirts",
					params.Name,
				)
			}

			if params.Code != "FORMAL-SHIRTS" {
				t.Fatalf(
					"Code = %q, want FORMAL-SHIRTS",
					params.Code,
				)
			}

			return domain.Category{
				ID:   params.ID,
				Name: params.Name,
				Code: params.Code,
			}, nil
		},
	}

	service := NewCategoryService(repo)

	_, err := service.Update(
		context.Background(),
		id,
		UpdateCategoryInput{
			Name: "  Formal Shirts  ",
			Code: "  FORMAL-SHIRTS  ",
		},
	)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}

func TestCategoryServiceSetActive(t *testing.T) {
	id := uuid.New()

	repo := &fakeCategoryRepository{
		setActiveFn: func(
			_ context.Context,
			gotID uuid.UUID,
			isActive bool,
		) (domain.Category, error) {
			if gotID != id {
				t.Fatal("incorrect category ID")
			}

			if isActive {
				t.Fatal("expected category to be deactivated")
			}

			return domain.Category{
				ID:       gotID,
				IsActive: isActive,
			}, nil
		},
	}

	service := NewCategoryService(repo)

	category, err := service.SetActive(
		context.Background(),
		id,
		false,
	)
	if err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}

	if category.IsActive {
		t.Fatal("category should be inactive")
	}
}
