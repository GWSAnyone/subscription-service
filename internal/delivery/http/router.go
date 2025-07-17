package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"github.com/subscription-service/internal/delivery/http/handler"
	"github.com/subscription-service/internal/delivery/http/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter создает новый маршрутизатор с настроенными эндпоинтами
func NewRouter(subscriptionHandler *handler.SubscriptionHandler) http.Handler {
	r := chi.NewRouter()

	// Подключаем глобальные middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recover)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// Настраиваем Swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json"), // URL к JSON-спецификации API
	))

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Endpoint для проверки работоспособности
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			log.Debug().Str("request_id", middleware.GetRequestID(r.Context())).Msg("Health check passed")
		})

		// Маршруты для подписок
		r.Route("/subscriptions", func(r chi.Router) {
			r.Post("/", subscriptionHandler.Create)
			r.Get("/", subscriptionHandler.List)
			r.Get("/{id}", subscriptionHandler.Get)
			r.Put("/{id}", subscriptionHandler.Update)
			r.Delete("/{id}", subscriptionHandler.Delete)
			r.Get("/calculate-cost", subscriptionHandler.CalculateTotalCost)
		})
	})

	return r
}
