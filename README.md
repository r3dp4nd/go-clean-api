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