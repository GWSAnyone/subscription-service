package subscription

import (
	"errors"
	"fmt"
	"time"
)

// Константы ошибок
var (
	// ErrSubscriptionNotFound возвращается когда подписка не найдена
	ErrSubscriptionNotFound = errors.New("subscription not found")

	// ErrInvalidInput возвращается при некорректных входных данных
	ErrInvalidInput = errors.New("invalid input")
)

// ParseMonthYear парсит строку формата MM-YYYY в time.Time
func ParseMonthYear(dateStr string) (time.Time, error) {
	parsedDate, err := time.Parse("01-2006", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %w", err)
	}
	return parsedDate, nil
}

// FormatMonthYear форматирует time.Time в строку MM-YYYY
func FormatMonthYear(t time.Time) string {
	return t.Format("01-2006")
}
