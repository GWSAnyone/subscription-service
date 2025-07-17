package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Структура для подписки
type DockerSubscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Популярные сервисы подписок
var dockerServices = []struct {
	Name        string
	MinPrice    int
	MaxPrice    int
	Description string
}{
	{"Netflix", 599, 1599, "Стриминговый сервис фильмов и сериалов"},
	{"Spotify Premium", 199, 499, "Музыкальный стриминговый сервис"},
	{"YouTube Premium", 299, 799, "Премиум-подписка на YouTube"},
	{"Apple Music", 169, 299, "Музыкальный стриминговый сервис от Apple"},
	{"Disney+", 399, 799, "Стриминговый сервис от Disney"},
	{"HBO Max", 599, 999, "Стриминговый сервис от HBO"},
	{"Amazon Prime", 399, 899, "Премиум-подписка Amazon"},
	{"Yandex Plus", 199, 399, "Подписка на сервисы Яндекса"},
	{"Kinopoisk HD", 299, 599, "Стриминговый сервис фильмов и сериалов"},
	{"PlayStation Plus", 599, 1299, "Подписка для игровой консоли PlayStation"},
	{"Xbox Game Pass", 699, 1499, "Подписка на игры для Xbox"},
	{"Adobe Creative Cloud", 1999, 4999, "Пакет программ для дизайна и творчества"},
	{"Microsoft 365", 499, 999, "Офисный пакет Microsoft"},
	{"Google One", 139, 999, "Расширенное облачное хранилище Google"},
	{"iCloud+", 149, 999, "Облачное хранилище Apple"},
	{"Notion Premium", 499, 999, "Расширенная версия приложения для заметок"},
	{"Telegram Premium", 299, 299, "Премиум-подписка Telegram"},
	{"Tinkoff Pro", 199, 199, "Премиум-подписка банка Тинькофф"},
	{"SberPrime", 199, 399, "Подписка на сервисы Сбера"},
	{"VK Combo", 199, 299, "Подписка на сервисы VK"},
}

// Генерация случайного пользователя
func dockerGenerateRandomUserID() uuid.UUID {
	// Создаем несколько фиксированных пользователей для более реалистичных данных
	users := []string{
		"60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"70702fee-3bf2-5832-be7f-8747e89b1dcb",
		"80803fee-4bf3-6943-cf8f-9858f9a2eddc",
		"90904fee-5bf4-7a54-df9f-a969fab3feed",
		"a0a05fee-6bf5-8b65-efaf-ba7afbc4ffee",
	}

	// Выбираем случайного пользователя из списка
	userIndex := rand.Intn(len(users))
	userID, _ := uuid.Parse(users[userIndex])
	return userID
}

// Генерация случайной даты начала подписки
func dockerGenerateRandomStartDate() time.Time {
	// Генерируем дату в пределах последних 2 лет
	now := time.Now()
	minTime := now.AddDate(-2, 0, 0)

	// Разница в секундах
	diff := now.Unix() - minTime.Unix()

	// Случайное время в этом диапазоне
	randomSecs := rand.Int63n(diff)
	randomTime := minTime.Add(time.Duration(randomSecs) * time.Second)

	// Устанавливаем только год и месяц, день всегда 1
	return time.Date(randomTime.Year(), randomTime.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// Генерация случайной даты окончания (может быть nil)
func dockerGenerateRandomEndDate(startDate time.Time) *time.Time {
	// С вероятностью 30% подписка будет бессрочной (без даты окончания)
	if rand.Float32() < 0.3 {
		return nil
	}

	// Иначе генерируем дату окончания от 1 до 24 месяцев после начала
	months := rand.Intn(24) + 1
	endDate := startDate.AddDate(0, months, 0)

	// Если дата окончания в будущем, с вероятностью 50% делаем её nil (активная подписка)
	if endDate.After(time.Now()) && rand.Float32() < 0.5 {
		return nil
	}

	return &endDate
}

// Генерация случайной подписки
func dockerGenerateRandomSubscription() DockerSubscription {
	// Выбираем случайный сервис
	service := dockerServices[rand.Intn(len(dockerServices))]

	// Генерируем случайную цену в диапазоне
	price := service.MinPrice
	if service.MaxPrice > service.MinPrice {
		price = service.MinPrice + rand.Intn(service.MaxPrice-service.MinPrice+1)
	}

	// Генерируем даты
	startDate := dockerGenerateRandomStartDate()
	endDate := dockerGenerateRandomEndDate(startDate)

	// Текущее время для created_at и updated_at
	now := time.Now()

	return DockerSubscription{
		ID:          uuid.New(),
		ServiceName: service.Name,
		Price:       price,
		UserID:      dockerGenerateRandomUserID(),
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func main() {
	// Инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Подключение к базе данных (используем имя хоста postgres для Docker)
	connStr := "postgres://postgres:postgres@postgres:5432/subscription_service?sslmode=disable"
	fmt.Println("Подключение к базе данных...")

	// Попытки подключения с повторами
	var db *sql.DB
	var err error
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Попытка %d: Не удалось открыть соединение: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = db.Ping()
		if err == nil {
			fmt.Println("Успешное подключение к базе данных")
			break
		}

		log.Printf("Попытка %d: Не удалось подключиться к базе данных: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных после %d попыток: %v", maxRetries, err)
	}
	defer db.Close()

	// Создание контекста с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Количество подписок для генерации
	subscriptionCount := 100
	fmt.Printf("Генерация %d случайных подписок...\n", subscriptionCount)

	// Генерация и вставка подписок
	for i := 0; i < subscriptionCount; i++ {
		subscription := dockerGenerateRandomSubscription()

		// SQL запрос для вставки
		query := `
			INSERT INTO subscriptions (
				id, service_name, price, user_id, start_date, end_date, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8
			)
		`

		_, err := db.ExecContext(
			ctx,
			query,
			subscription.ID,
			subscription.ServiceName,
			subscription.Price,
			subscription.UserID,
			subscription.StartDate,
			subscription.EndDate,
			subscription.CreatedAt,
			subscription.UpdatedAt,
		)

		if err != nil {
			log.Printf("Ошибка при вставке подписки %d: %v", i+1, err)
			continue
		}

		if (i+1)%10 == 0 {
			fmt.Printf("Добавлено %d из %d подписок\n", i+1, subscriptionCount)
		}
	}

	fmt.Printf("Успешно добавлено %d подписок в базу данных\n", subscriptionCount)
}
