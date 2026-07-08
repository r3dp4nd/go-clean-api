package server

import "net/http"

func registerAPIV1Routes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("/api/v1/ping", handler.handleAPIV1Ping)

	registerAPIV1ProductRoutes(mux, handler)
}
