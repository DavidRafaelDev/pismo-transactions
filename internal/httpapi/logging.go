package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/DavidRafaelDev/pismo-transactions/internal/securelog"
)

// maxLoggedBodyBytes caps how much of each body we buffer for the access log.
// Matches the per-handler MaxBytesReader so we never log more than the API
// would accept anyway.
const maxLoggedBodyBytes = 1 << 20 // 1 MiB

// statusRecorder wraps a ResponseWriter to capture the HTTP status code and
// the response body for the access log. The default ResponseWriter does not
// expose the status after WriteHeader; Write is intercepted to tee the body
// into a buffer while still forwarding to the real writer.
type statusRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// withLogging emits one structured log line per request with the request
// and response bodies. Sensitive fields (currently document_number) are
// redacted via securelog.MaskJSONBody before anything is written to the log.
func withLogging(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Buffer the request body so the handler can still read it. Capped
		// to avoid huge uploads blowing memory in the middleware before
		// the handler's own MaxBytesReader kicks in.
		var reqBody []byte
		if r.Body != nil {
			buf, _ := io.ReadAll(io.LimitReader(r.Body, maxLoggedBodyBytes+1))
			if len(buf) > maxLoggedBodyBytes {
				buf = buf[:maxLoggedBodyBytes]
			}
			reqBody = buf
			r.Body = io.NopCloser(bytes.NewReader(reqBody))
		}

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		logger.Info("http.request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_body", bodyLogValue(reqBody),
			"response_body", bodyLogValue(rec.body.Bytes()),
		)
	})
}

// bodyLogValue returns the body in a form slog serializes sensibly:
// json.RawMessage when the body parses as JSON (so it is emitted inline
// instead of as an escaped string), a plain string when it is not, and
// nil when empty (slog omits nothing — the field just serializes to null).
func bodyLogValue(body []byte) any {
	if len(body) == 0 {
		return nil
	}
	masked, isJSON := securelog.MaskJSONBody(body)
	if isJSON {
		return json.RawMessage(masked)
	}
	return string(masked)
}
