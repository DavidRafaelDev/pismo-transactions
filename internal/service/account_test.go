package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
)

// fakeAccountRepo is a hand-written in-memory stand-in for AccountRepository.
// It supports duplicate detection by document_number to mirror the MySQL
// unique index behavior we will implement in Etapa 6.
type fakeAccountRepo struct {
	byID      map[int64]domain.Account
	byDoc     map[string]int64
	nextID    int64
	createErr error
	findErr   error
}

func newFakeAccountRepo() *fakeAccountRepo {
	return &fakeAccountRepo{
		byID:  make(map[int64]domain.Account),
		byDoc: make(map[string]int64),
	}
}

func (r *fakeAccountRepo) Create(_ context.Context, acc domain.Account) (int64, error) {
	if r.createErr != nil {
		return 0, r.createErr
	}
	if _, exists := r.byDoc[acc.DocumentNumber]; exists {
		return 0, domain.ErrDuplicateAccount
	}
	r.nextID++
	acc.ID = r.nextID
	r.byID[r.nextID] = acc
	r.byDoc[acc.DocumentNumber] = r.nextID
	return r.nextID, nil
}

func (r *fakeAccountRepo) FindByID(_ context.Context, id int64) (domain.Account, error) {
	if r.findErr != nil {
		return domain.Account{}, r.findErr
	}
	acc, ok := r.byID[id]
	if !ok {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	return acc, nil
}

func TestAccountService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path assigns id", func(t *testing.T) {
		svc := NewAccountService(newFakeAccountRepo())
		acc, err := svc.Create(ctx, "12345678900")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if acc.ID <= 0 {
			t.Errorf("expected positive ID, got %d", acc.ID)
		}
		if acc.DocumentNumber != "12345678900" {
			t.Errorf("DocumentNumber: got %q", acc.DocumentNumber)
		}
	})

	t.Run("invalid document surfaces domain error", func(t *testing.T) {
		svc := NewAccountService(newFakeAccountRepo())
		_, err := svc.Create(ctx, "abc")
		if !errors.Is(err, domain.ErrInvalidDocumentNumber) {
			t.Fatalf("expected ErrInvalidDocumentNumber, got %v", err)
		}
	})

	t.Run("duplicate surfaces domain error via repo", func(t *testing.T) {
		svc := NewAccountService(newFakeAccountRepo())
		if _, err := svc.Create(ctx, "12345678900"); err != nil {
			t.Fatalf("seed failed: %v", err)
		}
		_, err := svc.Create(ctx, "12345678900")
		if !errors.Is(err, domain.ErrDuplicateAccount) {
			t.Fatalf("expected ErrDuplicateAccount, got %v", err)
		}
	})

	t.Run("infra error is wrapped and propagated", func(t *testing.T) {
		repo := newFakeAccountRepo()
		boom := errors.New("db exploded")
		repo.createErr = boom
		svc := NewAccountService(repo)
		_, err := svc.Create(ctx, "12345678900")
		if !errors.Is(err, boom) {
			t.Fatalf("expected wrapped boom, got %v", err)
		}
	})
}

func TestAccountService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path returns stored account", func(t *testing.T) {
		svc := NewAccountService(newFakeAccountRepo())
		created, err := svc.Create(ctx, "12345678900")
		if err != nil {
			t.Fatalf("seed failed: %v", err)
		}
		got, err := svc.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got.ID != created.ID || got.DocumentNumber != created.DocumentNumber {
			t.Errorf("got %+v, want %+v", got, created)
		}
	})

	t.Run("invalid id rejected before hitting repo", func(t *testing.T) {
		svc := NewAccountService(newFakeAccountRepo())
		_, err := svc.Get(ctx, 0)
		if !errors.Is(err, domain.ErrInvalidAccountID) {
			t.Fatalf("expected ErrInvalidAccountID, got %v", err)
		}
	})

	t.Run("not found surfaces domain error", func(t *testing.T) {
		svc := NewAccountService(newFakeAccountRepo())
		_, err := svc.Get(ctx, 999)
		if !errors.Is(err, domain.ErrAccountNotFound) {
			t.Fatalf("expected ErrAccountNotFound, got %v", err)
		}
	})

	t.Run("infra error is wrapped and propagated", func(t *testing.T) {
		repo := newFakeAccountRepo()
		boom := errors.New("db exploded")
		repo.findErr = boom
		svc := NewAccountService(repo)
		_, err := svc.Get(ctx, 1)
		if !errors.Is(err, boom) {
			t.Fatalf("expected wrapped boom, got %v", err)
		}
	})
}
