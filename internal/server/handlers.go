package server

import (
	"log/slog"
	"net/http"

	"github.com/r3dp4nd/go-clean-api/internal/product"
)

type Handler struct {
	logger       *slog.Logger
	productStore *product.Store
}

func NewHandler(logger *slog.Logger, productStore *product.Store) *Handler {
	return &Handler{
		logger:       logger,
		productStore: productStore,
	}
}

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeNotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r, http.MethodGet)
		return
	}

	response := HomeResponse{
		Message: "Welcome to go-clean-api",
		Status:  "running",
	}

	writeJSON(w, http.StatusOK, response)
}
