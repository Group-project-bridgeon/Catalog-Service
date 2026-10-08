package cache

import (
	"context"
	"time"

	memorycache "github.com/Group-project-bridgeon/Catalog-Service/internal/cache"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type UnitRepository struct {
	next repository.UnitRepository
	list *memorycache.TTL[[]domain.Unit]
}

var _ repository.UnitRepository = (*UnitRepository)(nil)

func NewUnitRepository(
	next repository.UnitRepository,
	ttl time.Duration,
) *UnitRepository {
	return &UnitRepository{
		next: next,
		list: memorycache.NewTTL[[]domain.Unit](ttl),
	}
}

func (r *UnitRepository) Create(
	ctx context.Context,
	params repository.CreateUnitParams,
) (domain.Unit, error) {
	value, err := r.next.Create(
		ctx,
		params,
	)
	if err != nil {
		return domain.Unit{}, err
	}

	r.list.Invalidate()

	return value, nil
}

func (r *UnitRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Unit, error) {
	return r.next.GetByID(ctx, id)
}

func (r *UnitRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Unit, error) {
	return r.next.GetByCode(ctx, code)
}

func (r *UnitRepository) GetByName(
	ctx context.Context,
	name string,
) (domain.Unit, error) {
	return r.next.GetByName(ctx, name)
}

func (r *UnitRepository) List(
	ctx context.Context,
) ([]domain.Unit, error) {
	if value, ok := r.list.Get(); ok {
		return append(
			[]domain.Unit(nil),
			value...,
		), nil
	}

	value, err := r.next.List(ctx)
	if err != nil {
		return nil, err
	}

	r.list.Set(
		append(
			[]domain.Unit(nil),
			value...,
		),
	)

	return value, nil
}

func (r *UnitRepository) Update(
	ctx context.Context,
	params repository.UpdateUnitParams,
) (domain.Unit, error) {
	value, err := r.next.Update(
		ctx,
		params,
	)
	if err != nil {
		return domain.Unit{}, err
	}

	r.list.Invalidate()

	return value, nil
}

func (r *UnitRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	active bool,
) (domain.Unit, error) {
	value, err := r.next.SetActive(
		ctx,
		id,
		active,
	)
	if err != nil {
		return domain.Unit{}, err
	}

	r.list.Invalidate()

	return value, nil
}
