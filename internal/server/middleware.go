package server

import (
	"log/slog"
	"net/http"
	"time"
)

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := newResponseRecorder(w)

		next.ServeHTTP(recorder, r)

		duration := time.Since(startedAt)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", recorder.statusCode,
			"bytes_written", recorder.bytesWritten,
			"duration", duration.String(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		}

		switch {
		case recorder.statusCode >= http.StatusInternalServerError:
			logger.Error("http request completed", attrs...)
		case recorder.statusCode >= http.StatusBadRequest:
			logger.Warn("http request completed", attrs...)
		default:
			logger.Info("http request completed", attrs...)
		}
	})
}
