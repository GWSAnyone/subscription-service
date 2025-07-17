package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/subscription-service/internal/domain/subscription"
)

// SubscriptionRepository реализует интерфейс repository.SubscriptionRepository
type SubscriptionRepository struct {
	db *sqlx.DB
}

// NewSubscriptionRepository создает новый экземпляр репозитория подписок
func NewSubscriptionRepository(db *sqlx.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create создает новую запись о подписке
func (r *SubscriptionRepository) Create(ctx context.Context, sub *subscription.Subscription) error {
	query := `INSERT INTO subscriptions 
			(id, service_name, price, user_id, start_date, end_date, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	sub.ID = uuid.New()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.CreatedAt,
		sub.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// Get возвращает подписку по ID
func (r *SubscriptionRepository) Get(ctx context.Context, id uuid.UUID) (*subscription.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
			FROM subscriptions WHERE id = $1`

	var sub subscription.Subscription
	err := r.db.GetContext(ctx, &sub, query, id)
	if err != nil {
		// Проверяем, является ли ошибка "no rows in result set"
		if err.Error() == "sql: no rows in result set" {
			return nil, subscription.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &sub, nil
}

// Update обновляет существующую подписку
func (r *SubscriptionRepository) Update(ctx context.Context, sub *subscription.Subscription) error {
	query := `UPDATE subscriptions SET 
			service_name = $1, price = $2, start_date = $3, end_date = $4, updated_at = $5 
			WHERE id = $6`

	sub.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.StartDate,
		sub.EndDate,
		sub.UpdatedAt,
		sub.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return subscription.ErrSubscriptionNotFound
	}

	return nil
}

// Delete удаляет подписку по ID
func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return subscription.ErrSubscriptionNotFound
	}

	return nil
}

// List возвращает список всех подписок
func (r *SubscriptionRepository) List(ctx context.Context) ([]*subscription.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
			FROM subscriptions`

	var subs []*subscription.Subscription
	if err := r.db.SelectContext(ctx, &subs, query); err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return subs, nil
}

// CalculateTotalCost рассчитывает общую стоимость подписок по фильтру
func (r *SubscriptionRepository) CalculateTotalCost(ctx context.Context, filter subscription.SubscriptionFilter) (int, error) {
	// Строим запрос с использованием именованных параметров для безопасности
	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE 1=1`
	params := map[string]interface{}{}

	// Безопасно добавляем фильтр по ID пользователя (если указан)
	if filter.UserID != nil {
		query += " AND user_id = :user_id"
		params["user_id"] = *filter.UserID
	}

	// Безопасно добавляем фильтр по названию сервиса (если указан)
	if filter.ServiceName != nil && *filter.ServiceName != "" {
		query += " AND service_name = :service_name"
		params["service_name"] = *filter.ServiceName
	}

	// Добавляем фильтр по периоду (подписка должна действовать в указанном периоде)
	query += " AND (start_date <= :end_period)"
	params["end_period"] = filter.EndPeriod

	query += " AND (end_date IS NULL OR end_date >= :start_period)"
	params["start_period"] = filter.StartPeriod

	// Выполняем запрос с именованными параметрами
	nstmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare named statement: %w", err)
	}
	defer nstmt.Close()

	var totalCost int
	if err := nstmt.GetContext(ctx, &totalCost, params); err != nil {
		return 0, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	return totalCost, nil
}
