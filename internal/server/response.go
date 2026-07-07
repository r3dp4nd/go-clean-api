package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

const (
	errorCodeNotFound         = "not_found"
	errorCodeMethodNotAllowed = "method_not_allowed"
	errorCodeInternal         = "internal_error"
)

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("error writing json response", "error", err)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string) {
	response := ErrorResponse{
		Error: APIError{
			Code:      code,
			Message:   message,
			RequestID: getRequestID(r.Context()),
		},
	}

	writeJSON(w, statusCode, response)
}

func writeMethodNotAllowed(w http.ResponseWriter, r *http.Request, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)

	writeError(
		w,
		r,
		http.StatusMethodNotAllowed,
		errorCodeMethodNotAllowed,
		"method not allowed",
	)
}

func writeNotFound(w http.ResponseWriter, r *http.Request) {
	writeError(
		w,
		r,
		http.StatusNotFound,
		errorCodeNotFound,
		"route not found",
	)
}

func writeInternalError(w http.ResponseWriter, r *http.Request) {
	writeError(
		w,
		r,
		http.StatusInternalServerError,
		errorCodeInternal,
		"internal server error",
	)
}
