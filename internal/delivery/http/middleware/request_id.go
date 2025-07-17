package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// RequestIDKey - ключ контекста для request ID
type requestIDKey struct{}

// RequestID возвращает middleware для добавления уникального ID к каждому запросу
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Добавляем request ID в заголовки ответа
		w.Header().Set("X-Request-ID", requestID)

		// Добавляем request ID в контекст запроса
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)

		// Добавляем request ID в логи
		logger := log.With().Str("request_id", requestID).Logger()
		ctx = logger.WithContext(ctx)

		// Логируем входящий запрос
		logger.Debug().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Str("remote_addr", r.RemoteAddr).
			Msg("Request received")

		// Продолжаем цепочку обработки
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID возвращает request ID из контекста
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
