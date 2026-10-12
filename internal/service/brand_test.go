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

type fakeBrandRepository struct {
	createFn    func(context.Context, repository.CreateBrandParams) (domain.Brand, error)
	getByIDFn   func(context.Context, uuid.UUID) (domain.Brand, error)
	getByCodeFn func(context.Context, string) (domain.Brand, error)
	getByNameFn func(context.Context, string) (domain.Brand, error)
	listFn      func(context.Context) ([]domain.Brand, error)
	updateFn    func(context.Context, repository.UpdateBrandParams) (domain.Brand, error)
	setActiveFn func(context.Context, uuid.UUID, bool) (domain.Brand, error)
}

func (f *fakeBrandRepository) Create(
	ctx context.Context,
	params repository.CreateBrandParams,
) (domain.Brand, error) {
	if f.createFn == nil {
		return domain.Brand{}, nil
	}

	return f.createFn(ctx, params)
}

func (f *fakeBrandRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Brand, error) {
	if f.getByIDFn == nil {
		return domain.Brand{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeBrandRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Brand, error) {
	if f.getByCodeFn == nil {
		return domain.Brand{}, nil
	}

	return f.getByCodeFn(ctx, code)
}

func (f *fakeBrandRepository) GetByName(
	ctx context.Context,
	name string,
) (domain.Brand, error) {
	if f.getByNameFn == nil {
		return domain.Brand{}, nil
	}

	return f.getByNameFn(ctx, name)
}

func (f *fakeBrandRepository) List(
	ctx context.Context,
) ([]domain.Brand, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakeBrandRepository) Update(
	ctx context.Context,
	params repository.UpdateBrandParams,
) (domain.Brand, error) {
	if f.updateFn == nil {
		return domain.Brand{}, nil
	}

	return f.updateFn(ctx, params)
}

func (f *fakeBrandRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Brand, error) {
	if f.setActiveFn == nil {
		return domain.Brand{}, nil
	}

	return f.setActiveFn(ctx, id, isActive)
}

func TestBrandServiceCreate(t *testing.T) {
	var captured repository.CreateBrandParams

	repo := &fakeBrandRepository{
		createFn: func(
			_ context.Context,
			params repository.CreateBrandParams,
		) (domain.Brand, error) {
			captured = params

			return domain.Brand{
				ID:       params.ID,
				Name:     params.Name,
				Code:     params.Code,
				IsActive: params.IsActive,
			}, nil
		},
	}

	service := NewBrandService(repo)

	brand, err := service.Create(
		context.Background(),
		CreateBrandInput{
			Name: "  Nike  ",
			Code: "  NIKE  ",
		},
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if captured.ID == uuid.Nil {
		t.Fatal("Create() did not generate brand UUID")
	}

	if captured.Name != "Nike" {
		t.Fatalf("Name = %q, want Nike", captured.Name)
	}

	if captured.Code != "NIKE" {
		t.Fatalf("Code = %q, want NIKE", captured.Code)
	}

	if !captured.IsActive {
		t.Fatal("new brand should be active")
	}

	if brand.ID != captured.ID {
		t.Fatal("returned brand ID does not match repository result")
	}
}

func TestBrandServiceCreateValidation(t *testing.T) {
	tests := []struct {
		name  string
		input CreateBrandInput
	}{
		{
			name: "empty name",
			input: CreateBrandInput{
				Name: "",
				Code: "NIKE",
			},
		},
		{
			name: "name too long",
			input: CreateBrandInput{
				Name: strings.Repeat("a", 101),
				Code: "NIKE",
			},
		},
		{
			name: "empty code",
			input: CreateBrandInput{
				Name: "Nike",
				Code: "",
			},
		},
		{
			name: "code too long",
			input: CreateBrandInput{
				Name: "Nike",
				Code: strings.Repeat("a", 51),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewBrandService(
				&fakeBrandRepository{},
			)

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

func TestBrandServiceCreateConflict(t *testing.T) {
	repo := &fakeBrandRepository{
		createFn: func(
			context.Context,
			repository.CreateBrandParams,
		) (domain.Brand, error) {
			return domain.Brand{}, repository.ErrConflict
		},
	}

	service := NewBrandService(repo)

	_, err := service.Create(
		context.Background(),
		CreateBrandInput{
			Name: "Nike",
			Code: "NIKE",
		},
	)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf(
			"Create() error = %v, want ErrConflict",
			err,
		)
	}
}

func TestBrandServiceGetByIDNotFound(t *testing.T) {
	repo := &fakeBrandRepository{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.Brand, error) {
			return domain.Brand{}, repository.ErrNotFound
		},
	}

	service := NewBrandService(repo)

	_, err := service.GetByID(
		context.Background(),
		uuid.New(),
	)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf(
			"GetByID() error = %v, want ErrNotFound",
			err,
		)
	}
}

func TestBrandServiceGetByCode(t *testing.T) {
	repo := &fakeBrandRepository{
		getByCodeFn: func(
			_ context.Context,
			code string,
		) (domain.Brand, error) {
			if code != "NIKE" {
				t.Fatalf("code = %q, want NIKE", code)
			}

			return domain.Brand{
				Name: "Nike",
				Code: code,
			}, nil
		},
	}

	service := NewBrandService(repo)

	_, err := service.GetByCode(
		context.Background(),
		"  NIKE  ",
	)
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
}

func TestBrandServiceGetByName(t *testing.T) {
	repo := &fakeBrandRepository{
		getByNameFn: func(
			_ context.Context,
			name string,
		) (domain.Brand, error) {
			if name != "Nike" {
				t.Fatalf("name = %q, want Nike", name)
			}

			return domain.Brand{
				Name: name,
				Code: "NIKE",
			}, nil
		},
	}

	service := NewBrandService(repo)

	_, err := service.GetByName(
		context.Background(),
		"  Nike  ",
	)
	if err != nil {
		t.Fatalf("GetByName() error = %v", err)
	}
}

func TestBrandServiceList(t *testing.T) {
	repo := &fakeBrandRepository{
		listFn: func(
			context.Context,
		) ([]domain.Brand, error) {
			return []domain.Brand{
				{Name: "Nike", Code: "NIKE"},
				{Name: "Adidas", Code: "ADIDAS"},
			}, nil
		},
	}

	service := NewBrandService(repo)

	brands, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(brands) != 2 {
		t.Fatalf(
			"List() length = %d, want 2",
			len(brands),
		)
	}
}

func TestBrandServiceUpdate(t *testing.T) {
	id := uuid.New()

	repo := &fakeBrandRepository{
		updateFn: func(
			_ context.Context,
			params repository.UpdateBrandParams,
		) (domain.Brand, error) {
			if params.ID != id {
				t.Fatal("incorrect brand ID")
			}

			if params.Name != "Nike Sports" {
				t.Fatalf(
					"Name = %q, want Nike Sports",
					params.Name,
				)
			}

			return domain.Brand{
				ID:   params.ID,
				Name: params.Name,
				Code: params.Code,
			}, nil
		},
	}

	service := NewBrandService(repo)

	_, err := service.Update(
		context.Background(),
		id,
		UpdateBrandInput{
			Name: "  Nike Sports  ",
			Code: "  NIKE-SPORTS  ",
		},
	)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}

func TestBrandServiceSetActive(t *testing.T) {
	id := uuid.New()

	repo := &fakeBrandRepository{
		setActiveFn: func(
			_ context.Context,
			gotID uuid.UUID,
			isActive bool,
		) (domain.Brand, error) {
			if gotID != id {
				t.Fatal("incorrect brand ID")
			}

			return domain.Brand{
				ID:       gotID,
				IsActive: isActive,
			}, nil
		},
	}

	service := NewBrandService(repo)

	brand, err := service.SetActive(
		context.Background(),
		id,
		false,
	)
	if err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}

	if brand.IsActive {
		t.Fatal("brand should be inactive")
	}
}
