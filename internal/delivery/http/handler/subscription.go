package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/subscription-service/internal/domain/subscription"
)

// SubscriptionHandler обрабатывает HTTP запросы связанные с подписками
type SubscriptionHandler struct {
	service   subscription.Service
	validator *validator.Validate
}

// NewSubscriptionHandler создает новый экземпляр обработчика подписок
func NewSubscriptionHandler(service subscription.Service) *SubscriptionHandler {
	return &SubscriptionHandler{
		service:   service,
		validator: validator.New(),
	}
}

// Create обрабатывает запрос на создание подписки
// @Summary Создать подписку
// @Description Создает новую запись о подписке
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body subscription.CreateSubscriptionRequest true "Данные для создания подписки"
// @Success 201 {object} subscription.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req subscription.CreateSubscriptionRequest

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode request body")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Валидируем запрос
	if err := h.validator.Struct(req); err != nil {
		log.Error().Err(err).Msg("Validation failed")
		respondWithError(w, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	// Создаем подписку
	sub, err := h.service.Create(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create subscription")
		if errors.Is(err, subscription.ErrInvalidInput) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to create subscription")
		return
	}

	respondWithJSON(w, http.StatusCreated, sub)
}

// Get обрабатывает запрос на получение подписки по ID
// @Summary Получить подписку
// @Description Получает информацию о подписке по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} subscription.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [get]
func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID format")
		respondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	sub, err := h.service.Get(r.Context(), id)
	if err != nil {
		log.Error().Err(err).Str("id", id.String()).Msg("Failed to get subscription")
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			respondWithError(w, http.StatusNotFound, "Subscription not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get subscription")
		return
	}

	respondWithJSON(w, http.StatusOK, sub)
}

// Update обрабатывает запрос на обновление подписки
// @Summary Обновить подписку
// @Description Обновляет информацию о подписке
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param request body subscription.UpdateSubscriptionRequest true "Данные для обновления подписки"
// @Success 200 {object} subscription.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID format")
		respondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	var req subscription.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode request body")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Валидируем запрос
	if err := h.validator.Struct(req); err != nil {
		log.Error().Err(err).Msg("Validation failed")
		respondWithError(w, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	sub, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		log.Error().Err(err).Str("id", id.String()).Msg("Failed to update subscription")
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			respondWithError(w, http.StatusNotFound, "Subscription not found")
			return
		}
		if errors.Is(err, subscription.ErrInvalidInput) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to update subscription")
		return
	}

	respondWithJSON(w, http.StatusOK, sub)
}

// Delete обрабатывает запрос на удаление подписки
// @Summary Удалить подписку
// @Description Удаляет подписку по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID format")
		respondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		log.Error().Err(err).Str("id", id.String()).Msg("Failed to delete subscription")
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			respondWithError(w, http.StatusNotFound, "Subscription not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List обрабатывает запрос на получение списка всех подписок
// @Summary Список подписок
// @Description Получает список всех подписок
// @Tags subscriptions
// @Accept json
// @Produce json
// @Success 200 {array} subscription.Subscription
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	subs, err := h.service.List(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to list subscriptions")
		respondWithError(w, http.StatusInternalServerError, "Failed to list subscriptions")
		return
	}

	respondWithJSON(w, http.StatusOK, subs)
}

// CalculateTotalCost обрабатывает запрос на подсчет общей стоимости подписок
// @Summary Рассчитать стоимость подписок
// @Description Рассчитывает суммарную стоимость всех подписок за выбранный период
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id query string false "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param start_period query string true "Начало периода (MM-YYYY)"
// @Param end_period query string true "Конец периода (MM-YYYY)"
// @Success 200 {object} subscription.TotalCostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/calculate-cost [get]
func (h *SubscriptionHandler) CalculateTotalCost(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры запроса
	var filter subscription.SubscriptionFilter

	// Парсим ID пользователя (опциональный)
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.Error().Err(err).Str("user_id", userIDStr).Msg("Invalid user ID format")
			respondWithError(w, http.StatusBadRequest, "Invalid user ID format")
			return
		}
		filter.UserID = &userID
	}

	// Название сервиса (опциональное)
	serviceName := r.URL.Query().Get("service_name")
	if serviceName != "" {
		filter.ServiceName = &serviceName
	}

	// Парсим период (обязательные параметры)
	startPeriodStr := r.URL.Query().Get("start_period")
	if startPeriodStr == "" {
		log.Error().Msg("Start period is required")
		respondWithError(w, http.StatusBadRequest, "Start period is required")
		return
	}

	endPeriodStr := r.URL.Query().Get("end_period")
	if endPeriodStr == "" {
		log.Error().Msg("End period is required")
		respondWithError(w, http.StatusBadRequest, "End period is required")
		return
	}

	// Конвертируем строки в time.Time
	startPeriod, err := subscription.ParseMonthYear(startPeriodStr)
	if err != nil {
		log.Error().Err(err).Str("start_period", startPeriodStr).Msg("Invalid start period format")
		respondWithError(w, http.StatusBadRequest, "Invalid start period format")
		return
	}

	endPeriod, err := subscription.ParseMonthYear(endPeriodStr)
	if err != nil {
		log.Error().Err(err).Str("end_period", endPeriodStr).Msg("Invalid end period format")
		respondWithError(w, http.StatusBadRequest, "Invalid end period format")
		return
	}

	// Проверяем, что конечная дата не раньше начальной
	if endPeriod.Before(startPeriod) {
		log.Error().Msg("End period cannot be before start period")
		respondWithError(w, http.StatusBadRequest, "End period cannot be before start period")
		return
	}

	filter.StartPeriod = startPeriod
	filter.EndPeriod = endPeriod

	// Вызываем сервис для расчета
	totalCost, err := h.service.CalculateTotalCost(r.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to calculate total cost")
		respondWithError(w, http.StatusInternalServerError, "Failed to calculate total cost")
		return
	}

	respondWithJSON(w, http.StatusOK, totalCost)
}

// ErrorResponse представляет структуру ответа с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// respondWithError отправляет JSON-ответ с ошибкой
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

// respondWithJSON отправляет JSON-ответ
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal JSON response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write response")
	}
}
