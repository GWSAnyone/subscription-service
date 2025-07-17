package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/subscription-service/internal/domain/subscription"
)

// SubscriptionService реализует сервис для работы с подписками
type SubscriptionService struct {
	repo subscription.Repository
}

// NewSubscriptionService создает новый экземпляр сервиса подписок
func NewSubscriptionService(repo subscription.Repository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

// Create создает новую подписку
func (s *SubscriptionService) Create(ctx context.Context, req subscription.CreateSubscriptionRequest) (*subscription.Subscription, error) {
	// Преобразуем строку с датой начала в time.Time
	startDate, err := subscription.ParseMonthYear(req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	// Если указана дата окончания, преобразуем её
	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		parsedEndDate, err := subscription.ParseMonthYear(*req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date: %w", err)
		}

		// Проверка, что дата окончания не раньше даты начала
		if parsedEndDate.Before(startDate) {
			return nil, fmt.Errorf("%w: end date cannot be before start date", subscription.ErrInvalidInput)
		}

		endDate = &parsedEndDate
	}

	// Создаем объект подписки
	sub := &subscription.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// Сохраняем в репозиторий
	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

// Get возвращает подписку по ID
func (s *SubscriptionService) Get(ctx context.Context, id uuid.UUID) (*subscription.Subscription, error) {
	sub, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

// Update обновляет существующую подписку
func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, req subscription.UpdateSubscriptionRequest) (*subscription.Subscription, error) {
	// Получаем текущую подписку
	sub, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription for update: %w", err)
	}

	// Обновляем поля, если они указаны в запросе
	if req.ServiceName != "" {
		sub.ServiceName = req.ServiceName
	}

	if req.Price != nil {
		sub.Price = *req.Price
	}

	if req.StartDate != "" {
		startDate, err := subscription.ParseMonthYear(req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date: %w", err)
		}
		sub.StartDate = startDate
	}

	if req.EndDate != nil {
		if *req.EndDate == "" {
			// Если передана пустая строка, удаляем дату окончания
			sub.EndDate = nil
		} else {
			// Иначе парсим новую дату окончания
			endDate, err := subscription.ParseMonthYear(*req.EndDate)
			if err != nil {
				return nil, fmt.Errorf("invalid end date: %w", err)
			}

			// Проверяем, что дата окончания не раньше даты начала
			if endDate.Before(sub.StartDate) {
				return nil, fmt.Errorf("%w: end date cannot be before start date", subscription.ErrInvalidInput)
			}

			sub.EndDate = &endDate
		}
	}

	// Обновляем в репозитории
	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return sub, nil
}

// Delete удаляет подписку по ID
func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

// List возвращает список всех подписок
func (s *SubscriptionService) List(ctx context.Context) ([]*subscription.Subscription, error) {
	subs, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	return subs, nil
}

// CalculateTotalCost рассчитывает общую стоимость подписок за период
func (s *SubscriptionService) CalculateTotalCost(ctx context.Context, filter subscription.SubscriptionFilter) (*subscription.TotalCostResponse, error) {
	totalCost, err := s.repo.CalculateTotalCost(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	return &subscription.TotalCostResponse{
		TotalCost: totalCost,
	}, nil
}
