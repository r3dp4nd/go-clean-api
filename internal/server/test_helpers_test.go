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

	mux := http.NewServeMux()

	handlers := NewHandler(logger, productStore)

	registerRoutes(mux, handlers)

	return requestIDMiddleware(
		loggingMiddleware(logger, mux),
	)
}
