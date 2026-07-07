package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	requestIDHeader    = "X-Request-Id"
	maxRequestIDLength = 128
)

type requestIDContextKey struct{}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := normalizeRequestID(r.Header.Get(requestIDHeader))

		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)

		w.Header().Set(requestIDHeader, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getRequestID(ctx context.Context) string {
	value, ok := ctx.Value(requestIDContextKey{}).(string)
	if !ok {
		return ""
	}

	return value
}

func normalizeRequestID(value string) string {
	value = strings.TrimSpace(value)

	if value == "" {
		return generateRequestID()
	}

	if len(value) > maxRequestIDLength {
		return generateRequestID()
	}

	if strings.ContainsAny(value, "\r\n") {
		return generateRequestID()
	}

	return value
}

func generateRequestID() string {
	var bytes [16]byte

	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(bytes[:])
}
