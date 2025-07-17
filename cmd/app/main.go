package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/subscription-service/configs"
	httpDelivery "github.com/subscription-service/internal/delivery/http"
	"github.com/subscription-service/internal/delivery/http/handler"
	"github.com/subscription-service/internal/repository/postgresql"
	"github.com/subscription-service/internal/usecase"
)

// @title Subscription Service API
// @version 1.0
// @description API for managing user subscriptions

// @host localhost:8080
// @BasePath /api/v1

func main() {
	// Инициализируем логгер
	setupLogger()

	// Загружаем конфигурацию
	config, err := configs.LoadConfig("")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Настраиваем логгер согласно конфигурации
	setupLoggerFromConfig(config.Logger)

	// Подключаемся к базе данных
	db, err := setupDatabase(config.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Применяем миграции
	if err := applyMigrations(db.DB, config.Database); err != nil {
		log.Fatal().Err(err).Msg("Failed to apply migrations")
	}

	// Инициализируем репозиторий
	subscriptionRepo := postgresql.NewSubscriptionRepository(db)

	// Инициализируем сервис
	subscriptionService := usecase.NewSubscriptionService(subscriptionRepo)

	// Инициализируем HTTP-обработчики
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)

	// Создаем маршрутизатор
	router := httpDelivery.NewRouter(subscriptionHandler)

	// Настраиваем HTTP-сервер
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Server.Port),
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Info().Int("port", config.Server.Port).Msg("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Ждем сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited properly")
}

// setupLogger настраивает базовый логгер
func setupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

// setupLoggerFromConfig настраивает логгер согласно конфигурации
func setupLoggerFromConfig(config configs.LoggerConfig) {
	// Настраиваем уровень логирования
	switch config.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Настраиваем формат вывода
	if config.Format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		log.Logger = log.Output(os.Stdout)
	}
}

// setupDatabase устанавливает соединение с базой данных
func setupDatabase(config configs.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg("Connected to database")
	return db, nil
}

// applyMigrations применяет миграции к базе данных
func applyMigrations(db *sql.DB, config configs.DatabaseConfig) error {
	log.Info().Str("path", config.MigrationsPath).Msg("Applying database migrations")

	// Создаем экземпляр драйвера для migrate
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: config.MigrationsTable,
	})
	if err != nil {
		return fmt.Errorf("failed to create migrations driver: %w", err)
	}

	// Создаем экземпляр migrate
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Применяем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Info().Msg("Migrations applied successfully")
	return nil
}
