package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Logger создает middleware для логирования HTTP-запросов
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем обертку для response writer, чтобы перехватить код статуса
		ww := NewResponseWriter(w)

		// Вызываем следующий обработчик в цепочке
		next.ServeHTTP(ww, r)

		// После обработки запроса логируем информацию
		duration := time.Since(start)

		// Формируем лог
		logger := log.Info()
		if ww.statusCode >= 400 {
			logger = log.Error()
		}

		logger.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.statusCode).
			Dur("duration", duration).
			Str("ip", r.RemoteAddr).
			Str("user-agent", r.UserAgent()).
			Msg("HTTP request")
	})
}

// ResponseWriter - обертка для http.ResponseWriter, которая сохраняет код статуса
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter создает новую обертку для http.ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

// WriteHeader сохраняет код статуса и вызывает оригинальный WriteHeader
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
