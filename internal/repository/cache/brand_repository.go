package cache

import (
	"context"
	"time"

	memorycache "github.com/Group-project-bridgeon/Catalog-Service/internal/cache"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type BrandRepository struct {
	next repository.BrandRepository
	list *memorycache.TTL[[]domain.Brand]
}

var _ repository.BrandRepository = (*BrandRepository)(nil)

func NewBrandRepository(
	next repository.BrandRepository,
	ttl time.Duration,
) *BrandRepository {
	return &BrandRepository{
		next: next,
		list: memorycache.NewTTL[[]domain.Brand](ttl),
	}
}

func (r *BrandRepository) Create(
	ctx context.Context,
	params repository.CreateBrandParams,
) (domain.Brand, error) {
	value, err := r.next.Create(
		ctx,
		params,
	)
	if err != nil {
		return domain.Brand{}, err
	}

	r.list.Invalidate()

	return value, nil
}

func (r *BrandRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Brand, error) {
	return r.next.GetByID(ctx, id)
}

func (r *BrandRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Brand, error) {
	return r.next.GetByCode(ctx, code)
}

func (r *BrandRepository) GetByName(
	ctx context.Context,
	name string,
) (domain.Brand, error) {
	return r.next.GetByName(ctx, name)
}

func (r *BrandRepository) List(
	ctx context.Context,
) ([]domain.Brand, error) {
	if value, ok := r.list.Get(); ok {
		return append(
			[]domain.Brand(nil),
			value...,
		), nil
	}

	value, err := r.next.List(ctx)
	if err != nil {
		return nil, err
	}

	r.list.Set(
		append(
			[]domain.Brand(nil),
			value...,
		),
	)

	return value, nil
}

func (r *BrandRepository) Update(
	ctx context.Context,
	params repository.UpdateBrandParams,
) (domain.Brand, error) {
	value, err := r.next.Update(
		ctx,
		params,
	)
	if err != nil {
		return domain.Brand{}, err
	}

	r.list.Invalidate()

	return value, nil
}

func (r *BrandRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	active bool,
) (domain.Brand, error) {
	value, err := r.next.SetActive(
		ctx,
		id,
		active,
	)
	if err != nil {
		return domain.Brand{}, err
	}

	r.list.Invalidate()

	return value, nil
}
