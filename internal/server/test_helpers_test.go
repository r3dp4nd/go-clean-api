package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/r3dp4nd/go-clean-api/internal/audit"
	"github.com/r3dp4nd/go-clean-api/internal/product"
)

type fakeReadinessChecker struct {
	err error
}

func (f fakeReadinessChecker) Ping(ctx context.Context) error {
	return f.err
}

func newTestHTTPHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	productStore := product.NewStore()
	auditStore := audit.NewMemoryStore()

	productService := product.NewServiceWithAuditor(productStore, auditStore)

	mux := http.NewServeMux()

	handlers := NewHandler(
		logger,
		productService,
		auditStore,
		fakeReadinessChecker{},
	)

	registerRoutes(mux, handlers)

	corsOptions := CORSOptions{
		Enabled: true,
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:4200",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Request-ID",
		},
		MaxAgeSeconds: 600,
	}

	return requestIDMiddleware(
		loggingMiddleware(
			logger,
			recoveryMiddleware(
				logger,
				corsMiddleware(corsOptions, mux),
			),
		),
	)
}
