# Используем Multi-stage builds для уменьшения размера конечного образа

# Первый этап: сборка приложения
FROM golang:1.21-alpine as builder

# Установка необходимых зависимостей
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Создание директории для приложения
WORKDIR /app

# Копирование зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование кода приложения
COPY . .

# Запуск тестов
RUN go test -v ./...

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/subscription-service ./cmd/app

# Второй этап: создание минимального образа
FROM alpine:latest

# Установка необходимых зависимостей
RUN apk --no-cache add ca-certificates tzdata

# Копирование исполняемого файла из предыдущего этапа
COPY --from=builder /app/subscription-service /app/subscription-service

# Копирование миграций и конфигурационных файлов
COPY --from=builder /app/migrations /root/migrations
COPY --from=builder /app/configs /root/configs
COPY --from=builder /app/docs /root/docs

# Создание пользователя без привилегий
RUN adduser -D -g '' appuser
USER appuser

# Установка рабочей директории
WORKDIR /app

# Запуск приложения
CMD ["./subscription-service"] 