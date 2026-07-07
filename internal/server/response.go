package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error writing json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	response := ErrorResponse{
		Error: message,
	}

	writeJSON(w, statusCode, response)
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeNotFound(w http.ResponseWriter) {
	writeError(w, http.StatusNotFound, "route not found")
}
