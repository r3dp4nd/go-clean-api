package server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

const (
	errorCodeNotFound         = "not_found"
	errorCodeMethodNotAllowed = "method_not_allowed"
	errorCodeInvalidRequest   = "invalid_request"
	errorCodeInternal         = "internal_error"
)

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("error writing json response", "error", err)
	}
}

func readJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
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

func writeBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	writeError(
		w,
		r,
		http.StatusBadRequest,
		errorCodeInvalidRequest,
		message,
	)
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
