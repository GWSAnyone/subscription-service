package subscription

import (
	"context"

	"github.com/google/uuid"
)

// Service определяет интерфейс сервиса для работы с бизнес-логикой подписок
type Service interface {
	Create(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error)
	Get(ctx context.Context, id uuid.UUID) (*Subscription, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateSubscriptionRequest) (*Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*Subscription, error)
	CalculateTotalCost(ctx context.Context, filter SubscriptionFilter) (*TotalCostResponse, error)
}
