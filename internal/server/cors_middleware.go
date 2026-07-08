package server

import (
	"net/http"
	"strconv"
	"strings"
)

type CORSOptions struct {
	Enabled        bool
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAgeSeconds  int
}

func corsMiddleware(options CORSOptions, next http.Handler) http.Handler {
	allowedMethods := strings.Join(options.AllowedMethods, ", ")
	allowedHeaders := strings.Join(options.AllowedHeaders, ", ")
	maxAge := strconv.Itoa(options.MaxAgeSeconds)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !options.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		addVaryHeader(w, "Origin")
		addVaryHeader(w, "Access-Control-Request-Method")
		addVaryHeader(w, "Access-Control-Request-Headers")

		if !isOriginAllowed(origin, options.AllowedOrigins) {
			if isPreflightRequest(r) {
				writeError(
					w,
					r,
					http.StatusForbidden,
					errorCodeForbidden,
					"cors origin is not allowed",
				)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		if allowsWildcardOrigin(options.AllowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		if isPreflightRequest(r) {
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.Header().Set("Access-Control-Max-Age", maxAge)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isPreflightRequest(r *http.Request) bool {
	return r.Method == http.MethodOptions &&
		strings.TrimSpace(r.Header.Get("Access-Control-Request-Method")) != ""
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		allowedOrigin = strings.TrimSpace(allowedOrigin)

		if allowedOrigin == "*" {
			return true
		}

		if allowedOrigin == origin {
			return true
		}
	}

	return false
}

func allowsWildcardOrigin(allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if strings.TrimSpace(allowedOrigin) == "*" {
			return true
		}
	}

	return false
}

func addVaryHeader(w http.ResponseWriter, value string) {
	existingValue := w.Header().Get("Vary")
	if existingValue == "" {
		w.Header().Set("Vary", value)
		return
	}

	for _, part := range strings.Split(existingValue, ",") {
		if strings.EqualFold(strings.TrimSpace(part), value) {
			return
		}
	}

	w.Header().Set("Vary", existingValue+", "+value)
}
