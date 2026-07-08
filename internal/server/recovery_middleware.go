package server

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func recoveryMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			logger.Error(
				"panic recovered",
				"request_id", getRequestID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"panic", recovered,
				"stack", string(debug.Stack()),
			)

			writeInternalError(w, r)
		}()

		next.ServeHTTP(w, r)
	})
}
