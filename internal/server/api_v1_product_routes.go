package server

import "net/http"

func registerAPIV1ProductRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("/api/v1/products", handler.handleAPIV1Products)
	mux.HandleFunc("/api/v1/products/exists", handler.handleAPIV1ProductExists)
	mux.HandleFunc("/api/v1/products/sku/", handler.handleAPIV1ProductBySKU)
	mux.HandleFunc("/api/v1/products/", handler.handleAPIV1ProductByID)
}
