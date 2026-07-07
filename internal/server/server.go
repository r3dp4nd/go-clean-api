package server

import (
	"errors"
	"log"
	"net/http"
	"time"
)

type Options struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type Server struct {
	httpServer *http.Server
}

func New(options Options) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/ready", handleReady)

	httpServer := &http.Server{
		Addr:              options.Addr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: options.ReadHeaderTimeout,
		ReadTimeout:       options.ReadTimeout,
		WriteTimeout:      options.WriteTimeout,
		IdleTimeout:       options.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
	}
}

func (s *Server) Start() error {
	log.Printf("HTTP server listening on %s", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
