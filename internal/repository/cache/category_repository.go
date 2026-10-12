package cache

import (
	"context"
	"time"

	memorycache "github.com/Group-project-bridgeon/Catalog-Service/internal/cache"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type CategoryRepository struct {
	next repository.CategoryRepository
	list *memorycache.TTL[[]domain.Category]
}

var _ repository.CategoryRepository = (*CategoryRepository)(nil)

func NewCategoryRepository(
	next repository.CategoryRepository,
	ttl time.Duration,
) *CategoryRepository {
	return &CategoryRepository{
		next: next,
		list: memorycache.NewTTL[[]domain.Category](ttl),
	}
}

func (r *CategoryRepository) Create(
	ctx context.Context,
	params repository.CreateCategoryParams,
) (domain.Category, error) {
	value, err := r.next.Create(
		ctx,
		params,
	)
	if err != nil {
		return domain.Category{}, err
	}

	r.list.Invalidate()

	return value, nil
}

func (r *CategoryRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Category, error) {
	return r.next.GetByID(ctx, id)
}

func (r *CategoryRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Category, error) {
	return r.next.GetByCode(ctx, code)
}

func (r *CategoryRepository) List(
	ctx context.Context,
) ([]domain.Category, error) {
	if value, ok := r.list.Get(); ok {
		return append(
			[]domain.Category(nil),
			value...,
		), nil
	}

	value, err := r.next.List(ctx)
	if err != nil {
		return nil, err
	}

	r.list.Set(
		append(
			[]domain.Category(nil),
			value...,
		),
	)

	return value, nil
}

func (r *CategoryRepository) Update(
	ctx context.Context,
	params repository.UpdateCategoryParams,
) (domain.Category, error) {
	value, err := r.next.Update(
		ctx,
		params,
	)
	if err != nil {
		return domain.Category{}, err
	}

	r.list.Invalidate()

	return value, nil
}

func (r *CategoryRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	active bool,
) (domain.Category, error) {
	value, err := r.next.SetActive(
		ctx,
		id,
		active,
	)
	if err != nil {
		return domain.Category{}, err
	}

	r.list.Invalidate()

	return value, nil
}
