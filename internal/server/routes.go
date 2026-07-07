package server

import "net/http"

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/ready", handleReady)
}
