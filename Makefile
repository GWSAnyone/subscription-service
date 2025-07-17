.PHONY: build run test test-coverage clean docker-build docker-run migrate swagger

# Переменные
APP_NAME = subscription-service
BUILD_DIR = build
DOCKER_IMAGE = subscription-service:latest
MIGRATIONS_DIR = migrations
SWAGGER_DIR = api

# Основные команды
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/app

run: build
	./$(BUILD_DIR)/$(APP_NAME)

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Docker команды
docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run -p 8080:8080 $(DOCKER_IMAGE)

# Docker Compose команды
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Миграции
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "postgres://postgres:postgres@localhost:5432/subscription_service?sslmode=disable" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "postgres://postgres:postgres@localhost:5432/subscription_service?sslmode=disable" down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

# Swagger
swagger:
	swag init -g cmd/app/main.go -o docs

# Вспомогательные команды
lint:
	golangci-lint run

help:
	@echo "Доступные команды:"
	@echo "  make build            - Собрать приложение"
	@echo "  make run              - Запустить приложение"
	@echo "  make test             - Запустить тесты"
	@echo "  make test-coverage    - Запустить тесты с отчетом о покрытии"
	@echo "  make clean            - Очистить сборочные артефакты"
	@echo "  make docker-build     - Собрать Docker-образ"
	@echo "  make docker-run       - Запустить Docker-контейнер"
	@echo "  make docker-up        - Запустить все контейнеры через docker-compose"
	@echo "  make docker-down      - Остановить все контейнеры docker-compose"
	@echo "  make docker-logs      - Просмотр логов контейнеров"
	@echo "  make migrate-up       - Применить миграции"
	@echo "  make migrate-down     - Откатить миграции"
	@echo "  make migrate-create   - Создать новую миграцию"
	@echo "  make swagger          - Сгенерировать Swagger-документацию"
	@echo "  make lint             - Запустить линтер" 