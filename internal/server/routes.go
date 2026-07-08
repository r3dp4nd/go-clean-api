package server

import "net/http"

func registerRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("/", handler.handleHome)
	mux.HandleFunc("/debug/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("manual panic test")
	})

	registerSystemRoutes(mux, handler)
	registerAPIV1Routes(mux, handler)
}
