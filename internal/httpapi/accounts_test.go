package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
)

type fakeAccountSvc struct {
	createFn func(ctx context.Context, doc string) (domain.Account, error)
	getFn    func(ctx context.Context, id int64) (domain.Account, error)
}

func (f *fakeAccountSvc) Create(ctx context.Context, doc string) (domain.Account, error) {
	return f.createFn(ctx, doc)
}
func (f *fakeAccountSvc) Get(ctx context.Context, id int64) (domain.Account, error) {
	return f.getFn(ctx, id)
}

func postAccount(h *AccountsHandler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	return rec
}

func TestAccountsHandler_Create(t *testing.T) {
	t.Run("201 with account_id and document_number", func(t *testing.T) {
		fake := &fakeAccountSvc{
			createFn: func(_ context.Context, doc string) (domain.Account, error) {
				return domain.Account{ID: 42, DocumentNumber: doc}, nil
			},
		}
		rec := postAccount(NewAccountsHandler(fake), `{"document_number":"12345678900"}`)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status: want 201, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, `"account_id":42`) {
			t.Errorf("body missing account_id: %s", body)
		}
		if !strings.Contains(body, `"document_number":"12345678900"`) {
			t.Errorf("body missing document_number: %s", body)
		}
	})

	t.Run("400 invalid_document_number from domain", func(t *testing.T) {
		fake := &fakeAccountSvc{
			createFn: func(_ context.Context, _ string) (domain.Account, error) {
				return domain.Account{}, domain.ErrInvalidDocumentNumber
			},
		}
		rec := postAccount(NewAccountsHandler(fake), `{"document_number":"abc"}`)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: want 400, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"code":"invalid_document_number"`) {
			t.Errorf("body: %s", rec.Body.String())
		}
	})

	t.Run("409 duplicate_account", func(t *testing.T) {
		fake := &fakeAccountSvc{
			createFn: func(_ context.Context, _ string) (domain.Account, error) {
				return domain.Account{}, domain.ErrDuplicateAccount
			},
		}
		rec := postAccount(NewAccountsHandler(fake), `{"document_number":"12345678900"}`)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status: want 409, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"code":"duplicate_account"`) {
			t.Errorf("body: %s", rec.Body.String())
		}
	})

	t.Run("400 invalid_body on malformed json", func(t *testing.T) {
		rec := postAccount(NewAccountsHandler(&fakeAccountSvc{}), `{not-json`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: want 400, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"code":"invalid_body"`) {
			t.Errorf("body: %s", rec.Body.String())
		}
	})

	t.Run("400 invalid_body on unknown field (DisallowUnknownFields)", func(t *testing.T) {
		rec := postAccount(NewAccountsHandler(&fakeAccountSvc{}),
			`{"document_number":"12345678900","extra":"boom"}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: want 400, got %d", rec.Code)
		}
	})

	t.Run("500 internal_error on unknown svc error", func(t *testing.T) {
		fake := &fakeAccountSvc{
			createFn: func(_ context.Context, _ string) (domain.Account, error) {
				return domain.Account{}, errors.New("boom")
			},
		}
		rec := postAccount(NewAccountsHandler(fake), `{"document_number":"12345678900"}`)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status: want 500, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"code":"internal_error"`) {
			t.Errorf("body: %s", rec.Body.String())
		}
	})
}

func TestAccountsHandler_Get(t *testing.T) {
	// r.PathValue only gets populated when the request is routed through a
	// ServeMux with a pattern that declares {id}, so tests need a mini router.
	setup := func(svc accountService) http.Handler {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /accounts/{id}", NewAccountsHandler(svc).Get)
		return mux
	}

	t.Run("200 with account body", func(t *testing.T) {
		fake := &fakeAccountSvc{
			getFn: func(_ context.Context, id int64) (domain.Account, error) {
				return domain.Account{ID: id, DocumentNumber: "12345678900"}, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/accounts/42", nil)
		rec := httptest.NewRecorder()
		setup(fake).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status: want 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"account_id":42`) {
			t.Errorf("body: %s", rec.Body.String())
		}
	})

	t.Run("404 account_not_found", func(t *testing.T) {
		fake := &fakeAccountSvc{
			getFn: func(_ context.Context, _ int64) (domain.Account, error) {
				return domain.Account{}, domain.ErrAccountNotFound
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/accounts/999", nil)
		rec := httptest.NewRecorder()
		setup(fake).ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status: want 404, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"code":"account_not_found"`) {
			t.Errorf("body: %s", rec.Body.String())
		}
	})

	t.Run("400 on non-numeric id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/accounts/abc", nil)
		rec := httptest.NewRecorder()
		setup(&fakeAccountSvc{}).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: want 400, got %d", rec.Code)
		}
	})

	t.Run("400 on non-positive id (zero)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/accounts/0", nil)
		rec := httptest.NewRecorder()
		setup(&fakeAccountSvc{}).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: want 400, got %d", rec.Code)
		}
	})
}
