package server

import "net/http"

func (h *Handler) handleAPIV1Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r, http.MethodGet)
		return
	}

	response := PingResponse{
		Message:   "pong",
		Version:   "v1",
		RequestID: getRequestID(r.Context()),
	}

	writeJSON(w, http.StatusOK, response)
}
