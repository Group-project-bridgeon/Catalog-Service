package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type ProductDetailRepository interface {
	Get(
		context.Context,
		uuid.UUID,
	) (domain.ProductDetail, error)
}
