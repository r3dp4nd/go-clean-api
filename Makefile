APP_NAME=go-clean-api
BUILD_DIR=bin
MAIN_PACKAGE=./cmd/api

APP_ENV?=development
HTTP_HOST?=
HTTP_PORT?=8080
LOG_LEVEL?=info
LOG_FORMAT?=json

.PHONY: help run build clean test fmt vet tidy

help:
	@echo "Comandos disponibles:"
	@echo "  make run     - Ejecuta la API en modo desarrollo"
	@echo "  make build   - Compila la aplicación"
	@echo "  make clean   - Elimina archivos generados"
	@echo "  make test    - Ejecuta los tests"
	@echo "  make fmt     - Formatea el código Go"
	@echo "  make vet     - Analiza problemas comunes en el código"
	@echo "  make tidy    - Ordena dependencias del módulo"

run:
	APP_NAME=$(APP_NAME) \
	APP_VERSION=$(APP_VERSION) \
	APP_ENV=$(APP_ENV) \
	HTTP_HOST=$(HTTP_HOST) \
	HTTP_PORT=$(HTTP_PORT) \
	LOG_LEVEL=$(LOG_LEVEL) \
	LOG_FORMAT=$(LOG_FORMAT) \
	go run $(MAIN_PACKAGE)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PACKAGE)

clean:
	rm -rf $(BUILD_DIR)
	rm -f app

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy