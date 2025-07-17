package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/subscription-service/internal/domain/subscription"
)

// MockRepository - мок для репозитория подписок
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, sub *subscription.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockRepository) Get(ctx context.Context, id uuid.UUID) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, sub *subscription.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context) ([]*subscription.Subscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*subscription.Subscription), args.Error(1)
}

func (m *MockRepository) CalculateTotalCost(ctx context.Context, filter subscription.SubscriptionFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

func TestSubscriptionService_Create(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSubscriptionService(mockRepo)
	ctx := context.Background()

	// Подготовка тестовых данных
	userID := uuid.New()
	createReq := subscription.CreateSubscriptionRequest{
		ServiceName: "Test Service",
		Price:       100,
		UserID:      userID,
		StartDate:   "07-2023",
	}

	t.Run("успешное создание подписки", func(t *testing.T) {
		// Настройка мока
		mockRepo.On("Create", ctx, mock.AnythingOfType("*subscription.Subscription")).Return(nil).Once()

		// Вызов тестируемого метода
		result, err := service.Create(ctx, createReq)

		// Проверки
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, createReq.ServiceName, result.ServiceName)
		assert.Equal(t, createReq.Price, result.Price)
		assert.Equal(t, createReq.UserID, result.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка репозитория", func(t *testing.T) {
		// Настройка мока
		repoErr := errors.New("database error")
		mockRepo.On("Create", ctx, mock.AnythingOfType("*subscription.Subscription")).Return(repoErr).Once()

		// Вызов тестируемого метода
		result, err := service.Create(ctx, createReq)

		// Проверки
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create subscription")
		mockRepo.AssertExpectations(t)
	})

	t.Run("неверный формат даты", func(t *testing.T) {
		// Подготовка запроса с неверным форматом даты
		invalidReq := createReq
		invalidReq.StartDate = "invalid-date"

		// Вызов тестируемого метода
		result, err := service.Create(ctx, invalidReq)

		// Проверки
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid start date")
	})
}

func TestSubscriptionService_CalculateTotalCost(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSubscriptionService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	serviceName := "Test Service"
	startPeriod, _ := time.Parse("01-2006", "01-2023")
	endPeriod, _ := time.Parse("01-2006", "12-2023")

	filter := subscription.SubscriptionFilter{
		UserID:      &userID,
		ServiceName: &serviceName,
		StartPeriod: startPeriod,
		EndPeriod:   endPeriod,
	}

	t.Run("успешный расчет стоимости", func(t *testing.T) {
		// Настройка мока
		expectedCost := 1200
		mockRepo.On("CalculateTotalCost", ctx, filter).Return(expectedCost, nil).Once()

		// Вызов тестируемого метода
		result, err := service.CalculateTotalCost(ctx, filter)

		// Проверки
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedCost, result.TotalCost)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка репозитория", func(t *testing.T) {
		// Настройка мока
		repoErr := errors.New("database error")
		mockRepo.On("CalculateTotalCost", ctx, filter).Return(0, repoErr).Once()

		// Вызов тестируемого метода
		result, err := service.CalculateTotalCost(ctx, filter)

		// Проверки
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to calculate total cost")
		mockRepo.AssertExpectations(t)
	})
}
