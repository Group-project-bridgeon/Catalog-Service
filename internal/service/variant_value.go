package service

import (
	"context"
	"fmt"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/pagination"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type CreateVariantValueInput struct {
	VariantID        uuid.UUID
	AttributeID      uuid.UUID
	AttributeValueID uuid.UUID
}

type UpdateVariantValueInput struct {
	AttributeValueID uuid.UUID
}

type ListVariantValuesInput struct {
	Page      PageInput
	VariantID *uuid.UUID
}

type variantValueCursor struct {
	VariantID   uuid.UUID `json:"variant_id"`
	AttributeID uuid.UUID `json:"attribute_id"`
}

type VariantValueService struct {
	repository repository.VariantValueRepository
}

func NewVariantValueService(
	repository repository.VariantValueRepository,
) *VariantValueService {
	return &VariantValueService{
		repository: repository,
	}
}

func (s *VariantValueService) Create(
	ctx context.Context,
	input CreateVariantValueInput,
) (domain.VariantValue, error) {
	if err := validateVariantValueIDs(
		input.VariantID,
		input.AttributeID,
		input.AttributeValueID,
	); err != nil {
		return domain.VariantValue{}, err
	}

	exists, err := s.repository.CombinationExists(
		ctx,
		input.VariantID,
		input.AttributeID,
		input.AttributeValueID,
	)
	if err != nil {
		return domain.VariantValue{},
			mapRepositoryError(
				"check variant combination",
				err,
			)
	}

	if exists {
		return domain.VariantValue{}, fmt.Errorf(
			"variant attribute combination already exists: %w",
			ErrConflict,
		)
	}

	result, err := s.repository.Create(
		ctx,
		repository.CreateVariantValueParams{
			VariantID:        input.VariantID,
			AttributeID:      input.AttributeID,
			AttributeValueID: input.AttributeValueID,
		},
	)
	if err != nil {
		return domain.VariantValue{},
			mapRepositoryError(
				"create variant value",
				err,
			)
	}

	return result, nil
}

func (s *VariantValueService) GetByVariantAndAttribute(
	ctx context.Context,
	variantID uuid.UUID,
	attributeID uuid.UUID,
) (domain.VariantValue, error) {
	if variantID == uuid.Nil {
		return domain.VariantValue{}, fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	if attributeID == uuid.Nil {
		return domain.VariantValue{}, fmt.Errorf(
			"attribute id is required: %w",
			ErrInvalidArgument,
		)
	}

	result, err := s.repository.GetByVariantAndAttribute(
		ctx,
		variantID,
		attributeID,
	)
	if err != nil {
		return domain.VariantValue{},
			mapRepositoryError(
				"get variant value",
				err,
			)
	}

	return result, nil
}

func (s *VariantValueService) List(
	ctx context.Context,
	input ListVariantValuesInput,
) (Page[domain.VariantValue], error) {
	page, err := normalizePageInput(input.Page)
	if err != nil {
		return Page[domain.VariantValue]{}, fmt.Errorf(
			"invalid pagination: %w",
			ErrInvalidArgument,
		)
	}

	var cursor variantValueCursor
	if page.Cursor != "" {
		if err := pagination.Decode(
			page.Cursor,
			&cursor,
		); err != nil {
			return Page[domain.VariantValue]{}, fmt.Errorf(
				"invalid cursor: %w",
				ErrInvalidArgument,
			)
		}
	}

	var cursorVariantID *uuid.UUID
	var cursorAttributeID *uuid.UUID

	if page.Cursor != "" {
		cursorVariantID = &cursor.VariantID
		cursorAttributeID = &cursor.AttributeID
	}

	values, err := s.repository.List(
		ctx,
		repository.ListVariantValuesParams{
			CursorVariantID:   cursorVariantID,
			CursorAttributeID: cursorAttributeID,
			VariantID:         input.VariantID,
			PageSize:          int32(page.Limit + 1),
		},
	)
	if err != nil {
		return Page[domain.VariantValue]{},
			mapRepositoryError(
				"list variant values",
				err,
			)
	}

	values, hasMore := trimPage(
		values,
		page.Limit,
	)

	var nextCursor *string
	if hasMore && len(values) > 0 {
		last := values[len(values)-1]
		encoded, err := pagination.Encode(
			variantValueCursor{
				VariantID:   last.VariantID,
				AttributeID: last.AttributeID,
			},
		)
		if err != nil {
			return Page[domain.VariantValue]{},
				fmt.Errorf(
					"encode variant value cursor: %w",
					err,
				)
		}
		nextCursor = &encoded
	}

	return Page[domain.VariantValue]{
		Items:      values,
		NextCursor: nextCursor,
	}, nil
}

func (s *VariantValueService) ListByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
	page PageInput,
) (Page[domain.VariantValue], error) {
	if variantID == uuid.Nil {
		return Page[domain.VariantValue]{}, fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	return s.List(
		ctx,
		ListVariantValuesInput{
			Page:      page,
			VariantID: &variantID,
		},
	)
}

func (s *VariantValueService) Update(
	ctx context.Context,
	variantID uuid.UUID,
	attributeID uuid.UUID,
	input UpdateVariantValueInput,
) (domain.VariantValue, error) {
	if err := validateVariantValueIDs(
		variantID,
		attributeID,
		input.AttributeValueID,
	); err != nil {
		return domain.VariantValue{}, err
	}

	exists, err := s.repository.CombinationExists(
		ctx,
		variantID,
		attributeID,
		input.AttributeValueID,
	)
	if err != nil {
		return domain.VariantValue{},
			mapRepositoryError(
				"check variant combination",
				err,
			)
	}

	if exists {
		return domain.VariantValue{}, fmt.Errorf(
			"variant attribute combination already exists: %w",
			ErrConflict,
		)
	}

	result, err := s.repository.Update(
		ctx,
		repository.UpdateVariantValueParams{
			VariantID:        variantID,
			AttributeID:      attributeID,
			AttributeValueID: input.AttributeValueID,
		},
	)
	if err != nil {
		return domain.VariantValue{},
			mapRepositoryError(
				"update variant value",
				err,
			)
	}

	return result, nil
}

func validateVariantValueIDs(
	variantID uuid.UUID,
	attributeID uuid.UUID,
	attributeValueID uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	if attributeID == uuid.Nil {
		return fmt.Errorf(
			"attribute id is required: %w",
			ErrInvalidArgument,
		)
	}

	if attributeValueID == uuid.Nil {
		return fmt.Errorf(
			"attribute value id is required: %w",
			ErrInvalidArgument,
		)
	}

	return nil
}
