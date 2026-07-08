# go-clean-api

Proyecto de aprendizaje profesional de Go orientado a backend, APIs, microservicios, Docker y Kubernetes.

## Objetivo

Construir paso a paso una API backend en Go aplicando buenas prácticas modernas de la industria.

## Stack inicial

- Go 1.26.x
- macOS ARM64
- Go Modules
- Librería estándar de Go

## Comandos básicos

Ejecutar:

```bash
go run main.go
```


---

## Comandos con Makefile

Ver comandos disponibles:

```bash
make help
```

## Endpoints iniciales

### Home

```bash
curl http://localhost:8080
```

## Configuración

La aplicación usa variables de entorno para configurar valores principales.

Variables disponibles:

```env
APP_NAME=go-clean-api
APP_VERSION=v0.1.0
APP_ENV=development

HTTP_HOST=
HTTP_PORT=8080
```

## Logging

La aplicación usa `log/slog` para logs estructurados.

Variables disponibles:

```env
LOG_LEVEL=info
LOG_FORMAT=json
```

## Logging HTTP

El middleware HTTP registra información estructurada de cada request.

Campos registrados:

```text
method
path
status_code
bytes_written
duration
remote_addr
user_agent
```

## Graceful shutdown

La aplicación captura señales del sistema para apagarse correctamente.

Señales soportadas:

```text
SIGINT  # CTRL + C
SIGTERM # Docker / Kubernetes
```

## Request ID

La API agrega un identificador único por request usando el header:

```http
X-Request-ID
```

## Formato de errores HTTP

La API responde errores con un formato estructurado:

```json
{
  "error": {
    "code": "not_found",
    "message": "route not found",
    "request_id": "test-request-123"
  }
}
```

## Handler con dependencias

La capa HTTP usa una estructura `Handler` para agrupar dependencias de los handlers.

```go
type Handler struct {
	logger *slog.Logger
}
```

## Rutas de sistema

La API separa las rutas técnicas de las rutas de negocio.

Rutas disponibles:

```text
GET /health
GET /ready
```

## API v1

La API pública versionada comienza bajo el prefijo:

```text
/api/v1
```

## Products API

Primer módulo de negocio de la API.

Por ahora usa almacenamiento en memoria.

### Listar productos

```bash
curl -i http://localhost:8080/api/v1/products
```

## Testing

La API usa `net/http/httptest` para probar handlers HTTP sin levantar un servidor real.

### Ejecutar tests

```bash
make test
```

## Tests del paquete product

El paquete `internal/product` contiene tests unitarios del store en memoria.

Casos cubiertos:

```text
Create
List
Get
Update
Delete
ErrNotFound
Context cancelado
Concurrencia básica
```

## Validación de requests

La API diferencia entre errores de request inválido y errores de validación.

### JSON inválido

Cuando el body no es JSON válido, la API responde:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "invalid request body",
    "request_id": "invalid-json-test"
  }
}
```

## Capa de servicio para Products

El módulo `Products` usa una separación por capas:

```text
HTTP Handler → Product Service → Product Repository
```

## Repository interface y fakes para tests

El módulo `Products` usa una interfaz `Repository` para desacoplar el service de la persistencia concreta.

```go
type Repository interface {
	List(ctx context.Context) ([]Product, error)
	Get(ctx context.Context, id string) (Product, error)
	Create(ctx context.Context, input CreateProductInput) (Product, error)
	Update(ctx context.Context, id string, input UpdateProductInput) (Product, error)
	Delete(ctx context.Context, id string) error
}
```

## Parsing de rutas REST

La API usa helpers internos para extraer parámetros desde rutas REST.

Archivo principal:

```text
internal/server/path_params.go
```

## Recovery middleware

La API usa un middleware de recuperación para capturar `panic` inesperados dentro de handlers HTTP.

Archivo principal:

```text
internal/server/recovery_middleware.go
```

## CORS

La API usa un middleware CORS configurable por variables de entorno.

Variables disponibles:

```env
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173,http://localhost:4200
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Request-ID
CORS_MAX_AGE_SECONDS=600
```

## Paginación de Products

El endpoint `GET /api/v1/products` soporta paginación por query params.

### Parámetros

```text
page      # número de página, mínimo 1
page_size # tamaño de página, mínimo 1, máximo 100
```