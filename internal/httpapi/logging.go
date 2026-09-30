package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder wraps a ResponseWriter to capture the HTTP status code.
// The default ResponseWriter does not expose the status after WriteHeader,
// so we intercept it here for the access log.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// withLogging is a middleware that emits one structured log line per request.
// It does NOT log request/response bodies to avoid leaking sensitive fields
// like document_number in logs.
func withLogging(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		logger.Info("http.request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
