package server

import "net/http"

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeNotFound(w)
		return
	}

	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	response := HomeResponse{
		Message: "Welcome to go-clean-api",
		Status:  "running",
	}

	writeJSON(w, http.StatusOK, response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	response := StatusResponse{
		Status: "ok",
	}

	writeJSON(w, http.StatusOK, response)
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	response := StatusResponse{
		Status: "ready",
	}

	writeJSON(w, http.StatusOK, response)
}
