package postgresql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/subscription-service/internal/domain/subscription"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDatabase(t *testing.T) (*sqlx.DB, func()) {
	ctx := context.Background()

	// Настройка контейнера PostgreSQL
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	// Запуск контейнера
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Получение хоста и порта
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Подключение к базе данных
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable", host, port.Port())
	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)

	// Создание тестовой таблицы
	_, err = db.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		
		CREATE TABLE IF NOT EXISTS subscriptions (
			id UUID PRIMARY KEY,
			service_name VARCHAR(255) NOT NULL,
			price INT NOT NULL CHECK (price > 0),
			user_id UUID NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);
		
		CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
		CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name);
		CREATE INDEX IF NOT EXISTS idx_subscriptions_date_range ON subscriptions(start_date, end_date);
	`)
	require.NoError(t, err)

	// Функция очистки
	cleanup := func() {
		db.Close()
		container.Terminate(ctx)
	}

	return db, cleanup
}

func TestSubscriptionRepository_CRUD(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	// Создаем тестовые данные
	userID := uuid.New()
	startDate := time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC)

	sub := &subscription.Subscription{
		ServiceName: "Test Service",
		Price:       100,
		UserID:      userID,
		StartDate:   startDate,
	}

	// Тест создания подписки
	t.Run("Create", func(t *testing.T) {
		err := repo.Create(ctx, sub)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, sub.ID)
	})

	// Тест получения подписки
	t.Run("Get", func(t *testing.T) {
		fetchedSub, err := repo.Get(ctx, sub.ID)
		assert.NoError(t, err)
		assert.Equal(t, sub.ID, fetchedSub.ID)
		assert.Equal(t, sub.ServiceName, fetchedSub.ServiceName)
		assert.Equal(t, sub.Price, fetchedSub.Price)
		assert.Equal(t, sub.UserID, fetchedSub.UserID)
	})

	// Тест обновления подписки
	t.Run("Update", func(t *testing.T) {
		sub.Price = 150
		sub.ServiceName = "Updated Service"

		err := repo.Update(ctx, sub)
		assert.NoError(t, err)

		// Проверяем, что данные обновились
		updatedSub, err := repo.Get(ctx, sub.ID)
		assert.NoError(t, err)
		assert.Equal(t, 150, updatedSub.Price)
		assert.Equal(t, "Updated Service", updatedSub.ServiceName)
	})

	// Тест списка подписок
	t.Run("List", func(t *testing.T) {
		// Создаем ещё одну подписку для проверки списка
		sub2 := &subscription.Subscription{
			ServiceName: "Another Service",
			Price:       200,
			UserID:      userID,
			StartDate:   startDate,
		}
		err := repo.Create(ctx, sub2)
		assert.NoError(t, err)

		// Получаем список всех подписок
		subs, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, subs, 2)
	})

	// Тест расчета стоимости
	t.Run("CalculateTotalCost", func(t *testing.T) {
		// Создаем фильтр для расчета стоимости
		filter := subscription.SubscriptionFilter{
			UserID:      &userID,
			StartPeriod: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndPeriod:   time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		}

		cost, err := repo.CalculateTotalCost(ctx, filter)
		assert.NoError(t, err)
		// Должно быть 150 (обновленная цена первой подписки) + 200 (вторая подписка) = 350
		assert.Equal(t, 350, cost)
	})

	// Тест удаления подписки
	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, sub.ID)
		assert.NoError(t, err)

		// Проверяем, что подписка удалена
		_, err = repo.Get(ctx, sub.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, subscription.ErrSubscriptionNotFound)
	})
}
