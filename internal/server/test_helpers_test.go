package server

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/r3dp4nd/go-clean-api/internal/product"
)

func newTestHTTPHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	productStore := product.NewStore()
	productService := product.NewService(productStore)

	mux := http.NewServeMux()

	handlers := NewHandler(logger, productService)

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
