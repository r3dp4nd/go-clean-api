package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSAllowedPreflightRequest(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/products", nil)
	request.Header.Set("Origin", "http://localhost:4200")
	request.Header.Set("Access-Control-Request-Method", "POST")
	request.Header.Set("Access-Control-Request-Headers", "Content-Type, X-Request-ID")
	request.Header.Set(requestIDHeader, "cors-preflight")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:4200" {
		t.Fatalf("expected allowed origin %q, got %q", "http://localhost:4200", got)
	}

	if got := responseRecorder.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "POST") {
		t.Fatalf("expected allowed methods to contain POST, got %q", got)
	}

	if got := responseRecorder.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Content-Type") {
		t.Fatalf("expected allowed headers to contain Content-Type, got %q", got)
	}

	if got := responseRecorder.Header().Get("Access-Control-Max-Age"); got != "600" {
		t.Fatalf("expected max age %q, got %q", "600", got)
	}

	if got := responseRecorder.Header().Get(requestIDHeader); got != "cors-preflight" {
		t.Fatalf("expected request id %q, got %q", "cors-preflight", got)
	}
}

func TestCORSActualRequest(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	request.Header.Set("Origin", "http://localhost:4200")
	request.Header.Set(requestIDHeader, "cors-actual-request")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:4200" {
		t.Fatalf("expected allowed origin %q, got %q", "http://localhost:4200", got)
	}
}

func TestCORSDisallowedPreflightRequest(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/products", nil)
	request.Header.Set("Origin", "https://evil.example.com")
	request.Header.Set("Access-Control-Request-Method", "POST")
	request.Header.Set(requestIDHeader, "cors-disallowed")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected empty allowed origin, got %q", got)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeForbidden {
		t.Fatalf("expected error code %q, got %q", errorCodeForbidden, response.Error.Code)
	}

	if response.Error.RequestID != "cors-disallowed" {
		t.Fatalf("expected request id %q, got %q", "cors-disallowed", response.Error.RequestID)
	}
}

func TestCORSRequestWithoutOriginPassesThrough(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	request.Header.Set(requestIDHeader, "no-origin")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS origin header, got %q", got)
	}
}
