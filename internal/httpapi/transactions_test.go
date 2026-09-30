package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
	"github.com/shopspring/decimal"
)

type fakeTxSvc struct {
	createFn func(
		ctx context.Context,
		accountID int64,
		opType domain.OperationType,
		amount decimal.Decimal,
	) (domain.Transaction, error)
}

func (f *fakeTxSvc) Create(
	ctx context.Context,
	accountID int64,
	opType domain.OperationType,
	amount decimal.Decimal,
) (domain.Transaction, error) {
	return f.createFn(ctx, accountID, opType, amount)
}

func postTransaction(h *TransactionsHandler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	return rec
}

func TestTransactionsHandler_Create(t *testing.T) {
	t.Run("201 with signed negative amount for purchase", func(t *testing.T) {
		fake := &fakeTxSvc{
			createFn: func(_ context.Context, accountID int64, opType domain.OperationType, amount decimal.Decimal) (domain.Transaction, error) {
				return domain.Transaction{
					ID: 1, AccountID: accountID, OperationTypeID: opType,
					Amount: amount.Neg(),
				}, nil
			},
		}
		rec := postTransaction(NewTransactionsHandler(fake),
			`{"account_id":1,"operation_type_id":1,"amount":"50.00"}`)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status: want 201, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, `"transaction_id":1`) {
			t.Errorf("body missing transaction_id: %s", body)
		}
		if !strings.Contains(body, `"amount":"-50`) {
			t.Errorf("body missing signed amount: %s", body)
		}
		if !strings.Contains(body, `"event_date"`) {
			t.Errorf("body missing event_date: %s", body)
		}
	})

	t.Run("201 with positive amount for credit voucher", func(t *testing.T) {
		fake := &fakeTxSvc{
			createFn: func(_ context.Context, accountID int64, opType domain.OperationType, amount decimal.Decimal) (domain.Transaction, error) {
				return domain.Transaction{
					ID: 2, AccountID: accountID, OperationTypeID: opType,
					Amount: amount, // credit voucher: no sign flip
				}, nil
			},
		}
		rec := postTransaction(NewTransactionsHandler(fake),
			`{"account_id":1,"operation_type_id":4,"amount":"60.00"}`)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status: want 201, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"amount":"60`) {
			t.Errorf("body missing positive amount: %s", rec.Body.String())
		}
	})

	t.Run("422 account_not_found (specific override for POST /transactions)", func(t *testing.T) {
		fake := &fakeTxSvc{
			createFn: func(_ context.Context, _ int64, _ domain.OperationType, _ decimal.Decimal) (domain.Transaction, error) {
				return domain.Transaction{}, domain.ErrAccountNotFound
			},
		}
		rec := postTransaction(NewTransactionsHandler(fake),
			`{"account_id":999,"operation_type_id":1,"amount":"10.00"}`)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status: want 422, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"code":"account_not_found"`) {
			t.Errorf("body: %s", rec.Body.String())
		}
	})

	t.Run("400 invalid_operation_type", func(t *testing.T) {
		fake := &fakeTxSvc{
			createFn: func(_ context.Context, _ int64, _ domain.OperationType, _ decimal.Decimal) (domain.Transaction, error) {
				return domain.Transaction{}, domain.ErrInvalidOperationType
			},
		}
		rec := postTransaction(NewTransactionsHandler(fake),
			`{"account_id":1,"operation_type_id":99,"amount":"10"}`)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: want 400, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"code":"invalid_operation_type"`) {
			t.Errorf("body: %s", rec.Body.String())
		}
	})

	t.Run("400 invalid_amount", func(t *testing.T) {
		fake := &fakeTxSvc{
			createFn: func(_ context.Context, _ int64, _ domain.OperationType, _ decimal.Decimal) (domain.Transaction, error) {
				return domain.Transaction{}, domain.ErrInvalidAmount
			},
		}
		rec := postTransaction(NewTransactionsHandler(fake),
			`{"account_id":1,"operation_type_id":1,"amount":"-10"}`)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: want 400, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"code":"invalid_amount"`) {
			t.Errorf("body: %s", rec.Body.String())
		}
	})

	t.Run("400 invalid_body on malformed json", func(t *testing.T) {
		rec := postTransaction(NewTransactionsHandler(&fakeTxSvc{}), `{bad`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: want 400, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"code":"invalid_body"`) {
			t.Errorf("body: %s", rec.Body.String())
		}
	})
}
