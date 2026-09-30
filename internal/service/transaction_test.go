package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
	"github.com/shopspring/decimal"
)

type fakeTxRepo struct {
	saved     []domain.Transaction
	nextID    int64
	createErr error
}

func newFakeTxRepo() *fakeTxRepo { return &fakeTxRepo{} }

func (r *fakeTxRepo) Create(_ context.Context, tx domain.Transaction) (int64, error) {
	if r.createErr != nil {
		return 0, r.createErr
	}
	r.nextID++
	tx.ID = r.nextID
	r.saved = append(r.saved, tx)
	return r.nextID, nil
}

func mustDecimal(t *testing.T, s string) decimal.Decimal {
	t.Helper()
	d, err := decimal.NewFromString(s)
	if err != nil {
		t.Fatalf("bad decimal literal %q: %v", s, err)
	}
	return d
}

func TestTransactionService_Create(t *testing.T) {
	ctx := context.Background()

	newServices := func() (*fakeAccountRepo, *fakeTxRepo, *TransactionService) {
		accRepo := newFakeAccountRepo()
		txRepo := newFakeTxRepo()
		svc := NewTransactionService(accRepo, txRepo)
		return accRepo, txRepo, svc
	}

	t.Run("happy path persists signed amount", func(t *testing.T) {
		accRepo, txRepo, svc := newServices()
		id, _ := accRepo.Create(ctx, domain.Account{DocumentNumber: "12345678900"})

		tx, err := svc.Create(ctx, id, domain.NormalPurchase, mustDecimal(t, "50.00"))
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if tx.ID <= 0 {
			t.Errorf("expected positive tx ID, got %d", tx.ID)
		}
		if !tx.Amount.Equal(mustDecimal(t, "-50.00")) {
			t.Errorf("Amount: want -50.00, got %s", tx.Amount)
		}
		if len(txRepo.saved) != 1 {
			t.Errorf("expected 1 saved tx, got %d", len(txRepo.saved))
		}
	})

	t.Run("credit voucher keeps positive sign", func(t *testing.T) {
		accRepo, _, svc := newServices()
		id, _ := accRepo.Create(ctx, domain.Account{DocumentNumber: "12345678900"})

		tx, err := svc.Create(ctx, id, domain.CreditVoucher, mustDecimal(t, "60.00"))
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if !tx.Amount.Equal(mustDecimal(t, "60.00")) {
			t.Errorf("Amount: want 60.00, got %s", tx.Amount)
		}
	})

	t.Run("account not found returns ErrAccountNotFound", func(t *testing.T) {
		_, _, svc := newServices()
		_, err := svc.Create(ctx, 999, domain.NormalPurchase, mustDecimal(t, "10.00"))
		if !errors.Is(err, domain.ErrAccountNotFound) {
			t.Fatalf("expected ErrAccountNotFound, got %v", err)
		}
	})

	t.Run("invalid op type surfaces domain error", func(t *testing.T) {
		accRepo, _, svc := newServices()
		id, _ := accRepo.Create(ctx, domain.Account{DocumentNumber: "12345678900"})

		_, err := svc.Create(ctx, id, 99, mustDecimal(t, "10.00"))
		if !errors.Is(err, domain.ErrInvalidOperationType) {
			t.Fatalf("expected ErrInvalidOperationType, got %v", err)
		}
	})

	t.Run("invalid amount surfaces domain error", func(t *testing.T) {
		accRepo, _, svc := newServices()
		id, _ := accRepo.Create(ctx, domain.Account{DocumentNumber: "12345678900"})

		_, err := svc.Create(ctx, id, domain.NormalPurchase, mustDecimal(t, "-1"))
		if !errors.Is(err, domain.ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("non-not-found account lookup error is wrapped", func(t *testing.T) {
		accRepo, _, svc := newServices()
		boom := errors.New("db exploded")
		accRepo.findErr = boom

		_, err := svc.Create(ctx, 1, domain.NormalPurchase, mustDecimal(t, "10.00"))
		if !errors.Is(err, boom) {
			t.Fatalf("expected wrapped boom, got %v", err)
		}
	})

	t.Run("tx repo error is wrapped and propagated", func(t *testing.T) {
		accRepo, txRepo, svc := newServices()
		id, _ := accRepo.Create(ctx, domain.Account{DocumentNumber: "12345678900"})
		boom := errors.New("insert failed")
		txRepo.createErr = boom

		_, err := svc.Create(ctx, id, domain.NormalPurchase, mustDecimal(t, "10.00"))
		if !errors.Is(err, boom) {
			t.Fatalf("expected wrapped boom, got %v", err)
		}
	})
}
