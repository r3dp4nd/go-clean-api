package server

import "net/http"

func registerRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("/", handler.handleHome)

	registerSystemRoutes(mux, handler)
}
