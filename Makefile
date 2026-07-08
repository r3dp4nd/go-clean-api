APP_NAME=go-clean-api
APP_VERSION=v0.1.0
BUILD_DIR=bin
MAIN_PACKAGE=./cmd/api

DOCKER_IMAGE?=go-clean-api
DOCKER_TAG?=local
DOCKER_PLATFORM?=linux/arm64

COMPOSE_FILE?=compose.yaml
COMPOSE_PROJECT_NAME?=go-clean-api

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

POSTGRES_PORT?=5432

DB_HOST?=localhost
DB_PORT?=5432
DB_NAME?=go_clean_api
DB_USER?=app
DB_PASSWORD?=app
DB_SSL_MODE?=disable

.PHONY: help run build clean test test-v test-cover test-race fmt vet tidy docker-build docker-run docker-stop docker-logs compose-build compose-up compose-up-d compose-down compose-down-v compose-logs compose-ps compose-db-logs compose-db-shell

help:
	@echo "Comandos disponibles:"
	@echo "  make run              - Ejecuta la API en modo desarrollo"
	@echo "  make build            - Compila la aplicación localmente"
	@echo "  make clean            - Elimina archivos generados"
	@echo "  make test             - Ejecuta los tests"
	@echo "  make test-v           - Ejecuta los tests en modo verbose"
	@echo "  make test-cover       - Ejecuta tests con cobertura"
	@echo "  make test-race        - Ejecuta tests con detector de race conditions"
	@echo "  make fmt              - Formatea el código Go"
	@echo "  make vet              - Analiza problemas comunes en el código"
	@echo "  make tidy             - Ordena dependencias del módulo"
	@echo "  make docker-build     - Construye la imagen Docker"
	@echo "  make docker-run       - Ejecuta la API en Docker"
	@echo "  make docker-stop      - Detiene y elimina el contenedor Docker"
	@echo "  make docker-logs      - Muestra logs del contenedor Docker"
	@echo "  make compose-build    - Construye servicios con Docker Compose"
	@echo "  make compose-up       - Levanta servicios con Docker Compose"
	@echo "  make compose-up-d     - Levanta servicios en segundo plano"
	@echo "  make compose-down     - Detiene servicios de Docker Compose"
	@echo "  make compose-down-v   - Detiene servicios y elimina volúmenes"
	@echo "  make compose-logs     - Muestra logs de Docker Compose"
	@echo "  make compose-ps       - Lista servicios de Docker Compose"
	@echo "  make compose-db-logs  - Muestra logs de PostgreSQL"
	@echo "  make compose-db-shell - Abre psql dentro del contenedor PostgreSQL"

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
	DB_HOST=$(DB_HOST) \
	DB_PORT=$(DB_PORT) \
	DB_NAME=$(DB_NAME) \
	DB_USER=$(DB_USER) \
	DB_PASSWORD=$(DB_PASSWORD) \
	DB_SSL_MODE=$(DB_SSL_MODE) \
	go run $(MAIN_PACKAGE)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PACKAGE)

clean:
	rm -rf $(BUILD_DIR)
	rm -f app
	rm -f coverage.out

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

docker-build:
	docker build \
		--platform $(DOCKER_PLATFORM) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		.

docker-run:
	docker run --rm \
		--name $(APP_NAME) \
		-p $(HTTP_PORT):8080 \
		-e APP_NAME=$(APP_NAME) \
		-e APP_VERSION=$(APP_VERSION) \
		-e APP_ENV=production \
		-e HTTP_HOST=0.0.0.0 \
		-e HTTP_PORT=8080 \
		-e HTTP_SHUTDOWN_TIMEOUT_SECONDS=$(HTTP_SHUTDOWN_TIMEOUT_SECONDS) \
		-e LOG_LEVEL=$(LOG_LEVEL) \
		-e LOG_FORMAT=$(LOG_FORMAT) \
		-e CORS_ENABLED=$(CORS_ENABLED) \
		-e CORS_ALLOWED_ORIGINS="$(CORS_ALLOWED_ORIGINS)" \
		-e CORS_ALLOWED_METHODS="$(CORS_ALLOWED_METHODS)" \
		-e CORS_ALLOWED_HEADERS="$(CORS_ALLOWED_HEADERS)" \
		-e CORS_MAX_AGE_SECONDS=$(CORS_MAX_AGE_SECONDS) \
		-e DB_HOST=$(DB_HOST) \
		-e DB_PORT=$(DB_PORT) \
		-e DB_NAME=$(DB_NAME) \
		-e DB_USER=$(DB_USER) \
		-e DB_PASSWORD=$(DB_PASSWORD) \
		-e DB_SSL_MODE=$(DB_SSL_MODE) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop:
	-docker rm -f $(APP_NAME)

docker-logs:
	docker logs -f $(APP_NAME)

compose-build:
	docker compose \
		-p $(COMPOSE_PROJECT_NAME) \
		-f $(COMPOSE_FILE) \
		build

compose-up:
	docker compose \
		-p $(COMPOSE_PROJECT_NAME) \
		-f $(COMPOSE_FILE) \
		up --build

compose-up-d:
	docker compose \
		-p $(COMPOSE_PROJECT_NAME) \
		-f $(COMPOSE_FILE) \
		up --build -d

compose-down:
	docker compose \
		-p $(COMPOSE_PROJECT_NAME) \
		-f $(COMPOSE_FILE) \
		down

compose-down-v:
	docker compose \
		-p $(COMPOSE_PROJECT_NAME) \
		-f $(COMPOSE_FILE) \
		down -v

compose-logs:
	docker compose \
		-p $(COMPOSE_PROJECT_NAME) \
		-f $(COMPOSE_FILE) \
		logs -f api

compose-ps:
	docker compose \
		-p $(COMPOSE_PROJECT_NAME) \
		-f $(COMPOSE_FILE) \
		ps

compose-db-logs:
	docker compose \
		-p $(COMPOSE_PROJECT_NAME) \
		-f $(COMPOSE_FILE) \
		logs -f postgres

compose-db-shell:
	docker compose \
		-p $(COMPOSE_PROJECT_NAME) \
		-f $(COMPOSE_FILE) \
		exec postgres psql -U $(DB_USER) -d $(DB_NAME)