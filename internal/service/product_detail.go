package service

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type ProductDetailService struct {
	repository repository.ProductDetailRepository
}

func NewProductDetailService(
	repository repository.ProductDetailRepository,
) *ProductDetailService {
	return &ProductDetailService{
		repository: repository,
	}
}

func (s *ProductDetailService) Get(
	ctx context.Context,
	productID uuid.UUID,
) (domain.ProductDetail, error) {
	if productID == uuid.Nil {
		return domain.ProductDetail{},
			ErrInvalidArgument
	}

	detail, err := s.repository.Get(
		ctx,
		productID,
	)
	if err != nil {
		return domain.ProductDetail{},
			mapRepositoryError(
				"get product detail",
				err,
			)
	}

	return detail, nil
}
