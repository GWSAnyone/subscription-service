package subscription

import (
	"context"

	"github.com/google/uuid"
)

// Repository определяет интерфейс для взаимодействия с хранилищем данных о подписках
type Repository interface {
	Create(ctx context.Context, subscription *Subscription) error
	Get(ctx context.Context, id uuid.UUID) (*Subscription, error)
	Update(ctx context.Context, subscription *Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*Subscription, error)
	CalculateTotalCost(ctx context.Context, filter SubscriptionFilter) (int, error)
}
