package server

import "net/http"

func registerRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("/", handler.handleHome)
	mux.HandleFunc("/health", handler.handleHealth)
	mux.HandleFunc("/ready", handler.handleReady)
}
