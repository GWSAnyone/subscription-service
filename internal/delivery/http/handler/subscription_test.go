package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/subscription-service/internal/domain/subscription"
)

// MockSubscriptionService мок для сервиса подписок
type MockSubscriptionService struct {
	mock.Mock
}

func (m *MockSubscriptionService) Create(ctx context.Context, req subscription.CreateSubscriptionRequest) (*subscription.Subscription, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) Get(ctx context.Context, id uuid.UUID) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) Update(ctx context.Context, id uuid.UUID, req subscription.UpdateSubscriptionRequest) (*subscription.Subscription, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubscriptionService) List(ctx context.Context) ([]*subscription.Subscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*subscription.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) CalculateTotalCost(ctx context.Context, filter subscription.SubscriptionFilter) (*subscription.TotalCostResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.TotalCostResponse), args.Error(1)
}

func TestSubscriptionHandler_Create(t *testing.T) {
	// Создаем мок сервиса
	mockService := new(MockSubscriptionService)
	handler := NewSubscriptionHandler(mockService)

	// Тестовые данные
	userID := uuid.New()
	now := time.Now()

	// Подготовка запроса
	reqBody := subscription.CreateSubscriptionRequest{
		ServiceName: "Test Service",
		Price:       100,
		UserID:      userID,
		StartDate:   "07-2023",
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Подготовка ответа от сервиса
	expectedResponse := &subscription.Subscription{
		ID:          uuid.New(),
		ServiceName: reqBody.ServiceName,
		Price:       reqBody.Price,
		UserID:      reqBody.UserID,
		StartDate:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Настройка мока
	mockService.On("Create", mock.Anything, reqBody).Return(expectedResponse, nil)

	// Создаем HTTP-запрос
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.Create(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseBody subscription.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.ID, responseBody.ID)
	assert.Equal(t, expectedResponse.ServiceName, responseBody.ServiceName)
	assert.Equal(t, expectedResponse.Price, responseBody.Price)

	// Проверяем, что мок был вызван
	mockService.AssertExpectations(t)
}

func TestSubscriptionHandler_Get(t *testing.T) {
	// Создаем мок сервиса
	mockService := new(MockSubscriptionService)
	handler := NewSubscriptionHandler(mockService)

	// Создаем маршрутизатор Chi для работы с URL-параметрами
	r := chi.NewRouter()
	r.Get("/api/v1/subscriptions/{id}", handler.Get)

	// Тестовые данные
	subscriptionID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	// Подготовка ответа от сервиса
	expectedSub := &subscription.Subscription{
		ID:          subscriptionID,
		ServiceName: "Test Service",
		Price:       100,
		UserID:      userID,
		StartDate:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Настройка мока
	mockService.On("Get", mock.Anything, subscriptionID).Return(expectedSub, nil)

	// Создаем HTTP-запрос
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/subscriptions/%s", subscriptionID), nil)

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Выполняем запрос
	r.ServeHTTP(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody subscription.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedSub.ID, responseBody.ID)
	assert.Equal(t, expectedSub.ServiceName, responseBody.ServiceName)
	assert.Equal(t, expectedSub.Price, responseBody.Price)

	// Проверяем, что мок был вызван
	mockService.AssertExpectations(t)
}

func TestSubscriptionHandler_CalculateTotalCost(t *testing.T) {
	// Создаем мок сервиса
	mockService := new(MockSubscriptionService)
	handler := NewSubscriptionHandler(mockService)

	// Тестовые данные
	userID := uuid.New()
	startPeriod := "01-2023"
	endPeriod := "12-2023"

	expectedTotalCost := &subscription.TotalCostResponse{
		TotalCost: 1200,
	}

	// Настройка мока
	mockService.On("CalculateTotalCost", mock.Anything, mock.AnythingOfType("subscription.SubscriptionFilter")).Return(expectedTotalCost, nil)

	// Создаем HTTP-запрос с параметрами запроса
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf(
		"/api/v1/subscriptions/calculate-cost?user_id=%s&start_period=%s&end_period=%s",
		userID,
		startPeriod,
		endPeriod,
	), nil)

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.CalculateTotalCost(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody subscription.TotalCostResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedTotalCost.TotalCost, responseBody.TotalCost)

	// Проверяем, что мок был вызван
	mockService.AssertExpectations(t)
}
