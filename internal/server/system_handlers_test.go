package server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/r3dp4nd/go-clean-api/internal/audit"
	"github.com/r3dp4nd/go-clean-api/internal/product"
)

func TestReadyReturnsOKWhenDatabaseIsReady(t *testing.T) {
	handler := newSystemTestHandler(fakeReadinessChecker{})

	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	request.Header.Set(requestIDHeader, "ready-ok")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response StatusResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "ready" {
		t.Fatalf("expected status %q, got %q", "ready", response.Status)
	}
}

func TestReadyReturnsServiceUnavailableWhenDatabaseIsNotReady(t *testing.T) {
	handler := newSystemTestHandler(fakeReadinessChecker{
		err: errors.New("database unavailable"),
	})

	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	request.Header.Set(requestIDHeader, "ready-failed")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, responseRecorder.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeServiceUnavailable {
		t.Fatalf("expected error code %q, got %q", errorCodeServiceUnavailable, response.Error.Code)
	}

	if response.Error.Message != "service is not ready" {
		t.Fatalf("expected message %q, got %q", "service is not ready", response.Error.Message)
	}

	if response.Error.RequestID != "ready-failed" {
		t.Fatalf("expected request id %q, got %q", "ready-failed", response.Error.RequestID)
	}
}

func newSystemTestHandler(readinessChecker ReadinessChecker) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	productStore := product.NewStore()
	auditStore := audit.NewMemoryStore()

	productService := product.NewServiceWithAuditor(productStore, auditStore)

	mux := http.NewServeMux()

	handlers := NewHandler(
		logger,
		productService,
		auditStore,
		readinessChecker,
	)

	registerRoutes(mux, handlers)

	return requestIDMiddleware(
		loggingMiddleware(
			logger,
			recoveryMiddleware(logger, mux),
		),
	)
}
