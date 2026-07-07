package server

import "net/http"

func registerSystemRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("/health", handler.handleHealth)
	mux.HandleFunc("/ready", handler.handleReady)
}
