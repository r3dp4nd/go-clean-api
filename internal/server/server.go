package server

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Options struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	Logger            *slog.Logger
}

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(options Options) *Server {
	mux := http.NewServeMux()

	registerRoutes(mux)

	httpServer := &http.Server{
		Addr:              options.Addr,
		Handler:           loggingMiddleware(options.Logger, mux),
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
