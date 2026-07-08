APP_NAME=go-clean-api
APP_VERSION=v0.1.0
BUILD_DIR=bin
MAIN_PACKAGE=./cmd/api

APP_ENV?=development
HTTP_HOST?=
HTTP_PORT?=8080
HTTP_SHUTDOWN_TIMEOUT_SECONDS?=10
LOG_LEVEL?=info
LOG_FORMAT?=json

CORS_ENABLED?=true
CORS_ALLOWED_ORIGINS?=http://localhost:3000,http://localhost:5173,http://localhost:4200
CORS_ALLOWED_METHODS?=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS?=Content-Type,Authorization,X-Request-ID
CORS_MAX_AGE_SECONDS?=600

.PHONY: help run build clean test test-v test-cover test-race fmt vet tidy

help:
	@echo "Comandos disponibles:"
	@echo "  make run        - Ejecuta la API en modo desarrollo"
	@echo "  make build      - Compila la aplicación"
	@echo "  make clean      - Elimina archivos generados"
	@echo "  make test       - Ejecuta los tests"
	@echo "  make test-v     - Ejecuta los tests en modo verbose"
	@echo "  make test-cover - Ejecuta tests con cobertura"
	@echo "  make test-race  - Ejecuta tests con detector de race conditions"
	@echo "  make fmt        - Formatea el código Go"
	@echo "  make vet        - Analiza problemas comunes en el código"
	@echo "  make tidy       - Ordena dependencias del módulo"

run:
	APP_NAME=$(APP_NAME) \
	APP_VERSION=$(APP_VERSION) \
	APP_ENV=$(APP_ENV) \
	HTTP_HOST=$(HTTP_HOST) \
	HTTP_PORT=$(HTTP_PORT) \
	HTTP_SHUTDOWN_TIMEOUT_SECONDS=$(HTTP_SHUTDOWN_TIMEOUT_SECONDS) \
	LOG_LEVEL=$(LOG_LEVEL) \
	LOG_FORMAT=$(LOG_FORMAT) \
	CORS_ENABLED=$(CORS_ENABLED) \
	CORS_ALLOWED_ORIGINS="$(CORS_ALLOWED_ORIGINS)" \
	CORS_ALLOWED_METHODS="$(CORS_ALLOWED_METHODS)" \
	CORS_ALLOWED_HEADERS="$(CORS_ALLOWED_HEADERS)" \
	CORS_MAX_AGE_SECONDS=$(CORS_MAX_AGE_SECONDS) \
	go run $(MAIN_PACKAGE)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PACKAGE)

clean:
	rm -rf $(BUILD_DIR)
	rm -f app

test:
	go test ./...

test-v:
	go test -v ./...

test-cover:
	go test -cover ./...

test-race:
	go test -race ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy