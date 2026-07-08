# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/go-clean-api \
    ./cmd/api

FROM alpine:3.22 AS runtime

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/go-clean-api /app/go-clean-api

ENV APP_NAME=go-clean-api
ENV APP_VERSION=v0.1.0
ENV APP_ENV=production

ENV HTTP_HOST=0.0.0.0
ENV HTTP_PORT=8080
ENV HTTP_SHUTDOWN_TIMEOUT_SECONDS=10

ENV LOG_LEVEL=info
ENV LOG_FORMAT=json

ENV CORS_ENABLED=true
ENV CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173,http://localhost:4200
ENV CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
ENV CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Request-ID
ENV CORS_MAX_AGE_SECONDS=600

EXPOSE 8080

ENTRYPOINT ["/app/go-clean-api"]