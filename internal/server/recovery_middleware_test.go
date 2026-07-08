package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoveryMiddlewareRecoversPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected test panic")
	})

	handler := requestIDMiddleware(
		loggingMiddleware(
			logger,
			recoveryMiddleware(logger, panicHandler),
		),
	)

	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	request.Header.Set(requestIDHeader, "panic-test")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get(requestIDHeader); got != "panic-test" {
		t.Fatalf("expected request id header %q, got %q", "panic-test", got)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeInternal {
		t.Fatalf("expected error code %q, got %q", errorCodeInternal, response.Error.Code)
	}

	if response.Error.Message != "internal server error" {
		t.Fatalf("expected error message %q, got %q", "internal server error", response.Error.Message)
	}

	if response.Error.RequestID != "panic-test" {
		t.Fatalf("expected request id %q, got %q", "panic-test", response.Error.RequestID)
	}
}

func TestRecoveryMiddlewarePassesThroughSuccessfulRequest(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, StatusResponse{
			Status: "ok",
		})
	})

	handler := requestIDMiddleware(
		loggingMiddleware(
			logger,
			recoveryMiddleware(logger, okHandler),
		),
	)

	request := httptest.NewRequest(http.MethodGet, "/ok", nil)
	request.Header.Set(requestIDHeader, "ok-test")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response StatusResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Fatalf("expected status %q, got %q", "ok", response.Status)
	}
}
