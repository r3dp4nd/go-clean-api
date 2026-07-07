package server

import (
	"encoding/json"
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

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	response := map[string]string{
		"message": "Welcome to go-clean-api",
		"status":  "running",
	}

	writeJSON(w, http.StatusOK, response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	response := map[string]string{
		"status": "ok",
	}

	writeJSON(w, http.StatusOK, response)
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	response := map[string]string{
		"status": "ready",
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error writing json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	response := map[string]string{
		"error": message,
	}

	writeJSON(w, statusCode, response)
}

func loggingMiddleware(next *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(startedAt))
	})
}
