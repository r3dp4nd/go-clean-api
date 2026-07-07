package server

import "net/http"

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

	response := StatusResponse{
		Status: "ready",
	}

	writeJSON(w, http.StatusOK, response)
}
