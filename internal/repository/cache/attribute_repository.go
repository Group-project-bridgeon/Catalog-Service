package cache

import (
	"context"
	"time"

	memorycache "github.com/Group-project-bridgeon/Catalog-Service/internal/cache"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type AttributeRepository struct {
	next repository.AttributeRepository
	list *memorycache.TTL[[]domain.Attribute]
}

var _ repository.AttributeRepository = (*AttributeRepository)(nil)

func NewAttributeRepository(
	next repository.AttributeRepository,
	ttl time.Duration,
) *AttributeRepository {
	return &AttributeRepository{
		next: next,
		list: memorycache.NewTTL[[]domain.Attribute](ttl),
	}
}

func (r *AttributeRepository) Create(
	ctx context.Context,
	params repository.CreateAttributeParams,
) (domain.Attribute, error) {
	value, err := r.next.Create(
		ctx,
		params,
	)
	if err != nil {
		return domain.Attribute{}, err
	}

	r.list.Invalidate()

	return value, nil
}

func (r *AttributeRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Attribute, error) {
	return r.next.GetByID(ctx, id)
}

func (r *AttributeRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Attribute, error) {
	return r.next.GetByCode(ctx, code)
}

func (r *AttributeRepository) GetByName(
	ctx context.Context,
	name string,
) (domain.Attribute, error) {
	return r.next.GetByName(ctx, name)
}

func (r *AttributeRepository) List(
	ctx context.Context,
) ([]domain.Attribute, error) {
	if value, ok := r.list.Get(); ok {
		return append(
			[]domain.Attribute(nil),
			value...,
		), nil
	}

	value, err := r.next.List(ctx)
	if err != nil {
		return nil, err
	}

	r.list.Set(
		append(
			[]domain.Attribute(nil),
			value...,
		),
	)

	return value, nil
}

func (r *AttributeRepository) Update(
	ctx context.Context,
	params repository.UpdateAttributeParams,
) (domain.Attribute, error) {
	value, err := r.next.Update(
		ctx,
		params,
	)
	if err != nil {
		return domain.Attribute{}, err
	}

	r.list.Invalidate()

	return value, nil
}
