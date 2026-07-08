package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIV1Ping(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	request.Header.Set(requestIDHeader, "test-ping")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get(requestIDHeader); got != "test-ping" {
		t.Fatalf("expected request id %q, got %q", "test-ping", got)
	}

	var response PingResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Message != "pong" {
		t.Fatalf("expected message %q, got %q", "pong", response.Message)
	}

	if response.Version != "v1" {
		t.Fatalf("expected version %q, got %q", "v1", response.Version)
	}

	if response.RequestID != "test-ping" {
		t.Fatalf("expected response request id %q, got %q", "test-ping", response.RequestID)
	}
}

func TestAPIV1PingMethodNotAllowed(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/ping", nil)
	request.Header.Set(requestIDHeader, "test-ping-405")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, got)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error.Code != errorCodeMethodNotAllowed {
		t.Fatalf("expected error code %q, got %q", errorCodeMethodNotAllowed, response.Error.Code)
	}

	if response.Error.RequestID != "test-ping-405" {
		t.Fatalf("expected request id %q, got %q", "test-ping-405", response.Error.RequestID)
	}
}
