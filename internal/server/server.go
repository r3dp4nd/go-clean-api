package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/r3dp4nd/go-clean-api/internal/product"
)

type Options struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	Logger            *slog.Logger
	ProductService    *product.Service
	ReadinessChecker  ReadinessChecker
	CORS              CORSOptions
}

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(options Options) *Server {
	mux := http.NewServeMux()

	handlers := NewHandler(
		options.Logger,
		options.ProductService,
		options.ReadinessChecker,
	)

	registerRoutes(mux, handlers)

	handlerChain := requestIDMiddleware(
		loggingMiddleware(
			options.Logger,
			recoveryMiddleware(
				options.Logger,
				corsMiddleware(options.CORS, mux),
			),
		),
	)

	httpServer := &http.Server{
		Addr:              options.Addr,
		Handler:           handlerChain,
		ReadHeaderTimeout: options.ReadHeaderTimeout,
		ReadTimeout:       options.ReadTimeout,
		WriteTimeout:      options.WriteTimeout,
		IdleTimeout:       options.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		logger:     options.Logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("http server listening", "addr", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("http server shutdown started")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	s.logger.Info("http server shutdown completed")

	return nil
}
