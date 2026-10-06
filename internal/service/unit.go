package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

const (
	maxUnitCodeLength = 30
	maxUnitNameLength = 100
)

type CreateUnitInput struct {
	Code string
	Name string
}

type UpdateUnitInput struct {
	Code string
	Name string
}

type UnitService struct {
	repository repository.UnitRepository
}

func NewUnitService(
	repository repository.UnitRepository,
) *UnitService {
	return &UnitService{
		repository: repository,
	}
}

func (s *UnitService) Create(
	ctx context.Context,
	input CreateUnitInput,
) (domain.Unit, error) {
	code, name, err := validateUnitFields(
		input.Code,
		input.Name,
	)
	if err != nil {
		return domain.Unit{}, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return domain.Unit{}, fmt.Errorf(
			"generate unit id: %w",
			err,
		)
	}

	unit, err := s.repository.Create(
		ctx,
		repository.CreateUnitParams{
			ID:       id,
			Code:     code,
			Name:     name,
			IsActive: true,
		},
	)
	if err != nil {
		return domain.Unit{}, mapRepositoryError(
			"create unit",
			err,
		)
	}

	return unit, nil
}

func (s *UnitService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Unit, error) {
	unit, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return domain.Unit{}, mapRepositoryError(
			"get unit",
			err,
		)
	}

	return unit, nil
}

func (s *UnitService) GetByCode(
	ctx context.Context,
	code string,
) (domain.Unit, error) {
	code = strings.TrimSpace(code)

	if code == "" {
		return domain.Unit{}, fmt.Errorf(
			"unit code is required: %w",
			ErrInvalidArgument,
		)
	}

	unit, err := s.repository.GetByCode(ctx, code)
	if err != nil {
		return domain.Unit{}, mapRepositoryError(
			"get unit by code",
			err,
		)
	}

	return unit, nil
}

func (s *UnitService) GetByName(
	ctx context.Context,
	name string,
) (domain.Unit, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return domain.Unit{}, fmt.Errorf(
			"unit name is required: %w",
			ErrInvalidArgument,
		)
	}

	unit, err := s.repository.GetByName(ctx, name)
	if err != nil {
		return domain.Unit{}, mapRepositoryError(
			"get unit by name",
			err,
		)
	}

	return unit, nil
}

func (s *UnitService) List(
	ctx context.Context,
) ([]domain.Unit, error) {
	units, err := s.repository.List(ctx)
	if err != nil {
		return nil, mapRepositoryError(
			"list units",
			err,
		)
	}

	return units, nil
}

func (s *UnitService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateUnitInput,
) (domain.Unit, error) {
	code, name, err := validateUnitFields(
		input.Code,
		input.Name,
	)
	if err != nil {
		return domain.Unit{}, err
	}

	unit, err := s.repository.Update(
		ctx,
		repository.UpdateUnitParams{
			ID:   id,
			Code: code,
			Name: name,
		},
	)
	if err != nil {
		return domain.Unit{}, mapRepositoryError(
			"update unit",
			err,
		)
	}

	return unit, nil
}

func (s *UnitService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Unit, error) {
	unit, err := s.repository.SetActive(
		ctx,
		id,
		isActive,
	)
	if err != nil {
		return domain.Unit{}, mapRepositoryError(
			"set unit active",
			err,
		)
	}

	return unit, nil
}

func validateUnitFields(
	code string,
	name string,
) (string, string, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)

	if code == "" {
		return "", "", fmt.Errorf(
			"unit code is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(code) > maxUnitCodeLength {
		return "", "", fmt.Errorf(
			"unit code exceeds %d characters: %w",
			maxUnitCodeLength,
			ErrInvalidArgument,
		)
	}

	if name == "" {
		return "", "", fmt.Errorf(
			"unit name is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(name) > maxUnitNameLength {
		return "", "", fmt.Errorf(
			"unit name exceeds %d characters: %w",
			maxUnitNameLength,
			ErrInvalidArgument,
		)
	}

	return code, name, nil
}
