# Makefile для geoservice

.PHONY: help install swagger-gen swagger-clean build run docker-build docker-run test clean

# Переменные
APP_NAME=geoservice
DOCKER_IMAGE=$(APP_NAME):latest
PORT=8080

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  install       - Установка зависимостей"
	@echo "  swagger-gen   - Генерация Swagger документации"
	@echo "  swagger-clean - Очистка документации"
	@echo "  build         - Сборка приложения"
	@echo "  run           - Запуск приложения локально"
	@echo "  test          - Запуск тестов"
	@echo "  docker-build  - Сборка Docker образа"
	@echo "  docker-run    - Запуск в Docker"
	@echo "  clean         - Очистка временных файлов"

# Установка зависимостей
install:
	@echo "📦 Установка зависимостей..."
	go mod download
	go mod tidy
	@echo "✅ Зависимости установлены"

# Установка swag CLI если не установлен
install-swag:
	@which swag > /dev/null || (echo "📦 Установка swag CLI..." && go install github.com/swaggo/swag/cmd/swag@latest)

# Генерация Swagger документации
swagger-gen: install-swag
	@echo "📚 Генерация Swagger документации..."
	swag init -g main.go --parseDependency --parseInternal
	@echo "✅ Документация сгенерирована в папке docs/"

# Очистка документации
swagger-clean:
	@echo "🧹 Очистка документации..."
	rm -rf docs/
	@echo "✅ Документация удалена"

# Полная пересборка документации
swagger-rebuild: swagger-clean swagger-gen

# Сборка приложения
build: swagger-gen
	@echo "🔨 Сборка приложения..."
	go build -o bin/$(APP_NAME) main.go
	@echo "✅ Приложение собрано: bin/$(APP_NAME)"

# Запуск приложения локально
run: swagger-gen
	@echo "Запуск $(APP_NAME)..."
	@echo "Swagger UI: http://localhost:$(PORT)/swagger"
	@echo "ReDoc: http://localhost:$(PORT)/docs"
	@echo "Health: http://localhost:$(PORT)/health"
	go run main.go

# Сборка Docker образа
docker-build: swagger-gen
	@echo "🐳 Сборка Docker образа..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "✅ Docker образ собран: $(DOCKER_IMAGE)"

# Запуск в Docker
docker-run: docker-build
	@echo "🐳 Запуск в Docker..."
	@echo "Swagger UI: http://localhost:$(PORT)/swagger"
	docker run -p $(PORT):$(PORT) --env-file .env $(DOCKER_IMAGE)

# Остановка docker-compose
docker-compose-down:
	@echo "🛑 Остановка docker-compose..."
	docker-compose down
	@echo "✅ Сервисы остановлены"

# Очистка
clean:
	@echo "🧹 Очистка временных файлов..."
	rm -rf bin/
	rm -rf docs/
	go clean
	docker system prune -f
	@echo "✅ Очистка завершена"

# Форматирование кода
fmt:
	@echo "Форматирование кода..."
	go fmt ./...
	@echo "✅ Код отформатирован"

# Линтинг (требует golangci-lint)
lint:
	@echo "🔍 Проверка кода..."
	golangci-lint run
	@echo "✅ Проверка завершена"

# Полный цикл разработки
dev: install swagger-gen fmt run

# Продакшн сборка
prod: install swagger-gen build docker-build