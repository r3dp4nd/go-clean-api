package server

import (
	"context"
	"net/http"
	"time"
)

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r, http.MethodGet)
		return
	}

	response := StatusResponse{
		Status: "ok",
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r, http.MethodGet)
		return
	}

	if h.readinessChecker != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.readinessChecker.Ping(ctx); err != nil {
			h.logger.Error(
				"readiness check failed",
				"error", err,
				"request_id", getRequestID(r.Context()),
			)

			writeError(
				w,
				r,
				http.StatusServiceUnavailable,
				errorCodeServiceUnavailable,
				"service is not ready",
			)
			return
		}
	}

	response := StatusResponse{
		Status: "ready",
	}

	writeJSON(w, http.StatusOK, response)
}
