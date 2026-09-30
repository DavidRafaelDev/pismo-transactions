package httpapi

import (
	"context"
	"net/http"
	"time"
)

// Pinger is defined here (the consumer) so tests can inject a fake without
// pulling in database/sql. *sql.DB satisfies it implicitly.
type Pinger interface {
	PingContext(ctx context.Context) error
}

func HealthHandler(p Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		if err := p.PingContext(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
