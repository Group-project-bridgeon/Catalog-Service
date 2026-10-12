package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCategoryRepo struct {
	mock.Mock
}

func (m *mockCategoryRepo) Create(ctx context.Context, params repository.CreateCategoryParams) (domain.Category, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(domain.Category), args.Error(1)
}

func (m *mockCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Category), args.Error(1)
}

func (m *mockCategoryRepo) GetByCode(ctx context.Context, code string) (domain.Category, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(domain.Category), args.Error(1)
}

func (m *mockCategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Category), args.Error(1)
}

func (m *mockCategoryRepo) Update(ctx context.Context, params repository.UpdateCategoryParams) (domain.Category, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(domain.Category), args.Error(1)
}

func (m *mockCategoryRepo) SetActive(ctx context.Context, id uuid.UUID, active bool) (domain.Category, error) {
	args := m.Called(ctx, id, active)
	return args.Get(0).(domain.Category), args.Error(1)
}

func TestCategoryRepository_CacheAndInvalidate(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockCategoryRepo)
	cached := NewCategoryRepository(mockRepo, time.Minute)

	cat1 := domain.Category{ID: uuid.New(), Name: "Electronics", Code: "ELEC"}
	catList := []domain.Category{cat1}

	// 1. First List calls underlying repo
	mockRepo.On("List", ctx).Return(catList, nil).Once()
	res1, err := cached.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, catList, res1)

	// 2. Second List before TTL returns cached value, mockRepo.List NOT called again
	res2, err := cached.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, catList, res2)
	mockRepo.AssertExpectations(t)

	// 3. Failed Create does NOT invalidate
	createParams := repository.CreateCategoryParams{ID: uuid.New(), Name: "Fashion", Code: "FASH"}
	mockRepo.On("Create", ctx, createParams).Return(domain.Category{}, errors.New("db error")).Once()
	_, err = cached.Create(ctx, createParams)
	assert.Error(t, err)

	// List should still be cached
	res3, err := cached.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, catList, res3)

	// 4. Successful Create invalidates cache
	cat2 := domain.Category{ID: createParams.ID, Name: createParams.Name, Code: createParams.Code}
	mockRepo.On("Create", ctx, createParams).Return(cat2, nil).Once()
	created, err := cached.Create(ctx, createParams)
	assert.NoError(t, err)
	assert.Equal(t, cat2, created)

	// Next List calls underlying repo again
	updatedList := []domain.Category{cat1, cat2}
	mockRepo.On("List", ctx).Return(updatedList, nil).Once()
	res4, err := cached.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, updatedList, res4)

	// 5. Successful Update invalidates cache
	updateParams := repository.UpdateCategoryParams{ID: cat1.ID, Name: "New Electronics", Code: "ELEC"}
	updatedCat1 := cat1
	updatedCat1.Name = "New Electronics"
	mockRepo.On("Update", ctx, updateParams).Return(updatedCat1, nil).Once()
	_, err = cached.Update(ctx, updateParams)
	assert.NoError(t, err)

	mockRepo.On("List", ctx).Return([]domain.Category{updatedCat1, cat2}, nil).Once()
	res5, err := cached.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []domain.Category{updatedCat1, cat2}, res5)

	// 6. SetActive invalidates cache
	mockRepo.On("SetActive", ctx, cat1.ID, false).Return(updatedCat1, nil).Once()
	_, err = cached.SetActive(ctx, cat1.ID, false)
	assert.NoError(t, err)

	mockRepo.On("List", ctx).Return([]domain.Category{cat2}, nil).Once()
	res6, err := cached.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []domain.Category{cat2}, res6)
	mockRepo.AssertExpectations(t)
}
