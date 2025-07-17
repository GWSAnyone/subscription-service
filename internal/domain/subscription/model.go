package subscription

import (
	"time"

	"github.com/google/uuid"
)

// Subscription представляет основную сущность подписки
type Subscription struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name" validate:"required"`
	Price       int        `json:"price" db:"price" validate:"required,min=1"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id" validate:"required"`
	StartDate   time.Time  `json:"start_date" db:"start_date" validate:"required"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateSubscriptionRequest представляет запрос на создание подписки
type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" validate:"required"`
	Price       int       `json:"price" validate:"required,min=1"`
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	StartDate   string    `json:"start_date" validate:"required"`
	EndDate     *string   `json:"end_date,omitempty"`
}

// UpdateSubscriptionRequest представляет запрос на обновление подписки
type UpdateSubscriptionRequest struct {
	ServiceName string  `json:"service_name,omitempty"`
	Price       *int    `json:"price,omitempty" validate:"omitempty,min=1"`
	StartDate   string  `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}

// SubscriptionFilter содержит параметры фильтрации для запросов
type SubscriptionFilter struct {
	UserID      *uuid.UUID `json:"user_id" form:"user_id"`
	ServiceName *string    `json:"service_name" form:"service_name"`
	StartPeriod time.Time  `json:"start_period" form:"start_period" validate:"required"`
	EndPeriod   time.Time  `json:"end_period" form:"end_period" validate:"required"`
}

// TotalCostResponse содержит результат расчета стоимости
type TotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}
