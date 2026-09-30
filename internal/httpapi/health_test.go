package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakePinger struct {
	err error
}

func (f fakePinger) PingContext(_ context.Context) error { return f.err }

// discardLogger builds a slog.Logger that swallows output — used in router
// tests so log lines don't pollute test output.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTestRouter builds a router with real handlers backed by empty fakes.
// Fakes of accountService / transactionService live in accounts_test.go
// and transactions_test.go (same package).
func newTestRouter(p Pinger) http.Handler {
	return NewRouter(
		p,
		NewAccountsHandler(&fakeAccountSvc{}),
		NewTransactionsHandler(&fakeTxSvc{}),
		discardLogger(),
	)
}

func TestHealthHandler_ReturnsOKWhenDBIsUp(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(fakePinger{err: nil})(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want %d, got %d", http.StatusOK, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type: want application/json, got %q", ct)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"status":"ok"`) {
		t.Errorf("body: want status=ok, got %s", body)
	}
}

func TestHealthHandler_Returns503WhenDBIsDown(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(fakePinger{err: errors.New("connection refused")})(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: want %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"status":"unhealthy"`) {
		t.Errorf("body: want status=unhealthy, got %s", body)
	}
}

func TestRouter_RoutesHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	newTestRouter(fakePinger{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRouter_RejectsWrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	newTestRouter(fakePinger{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status: want %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}
