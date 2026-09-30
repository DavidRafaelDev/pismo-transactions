//go:build integration

package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
)

func TestAccountRepository_Integration_Create(t *testing.T) {
	repo := NewAccountRepository(testDB)
	ctx := context.Background()

	t.Run("happy path assigns positive id", func(t *testing.T) {
		resetTables(t)
		acc, err := domain.NewAccount("12345678900")
		if err != nil {
			t.Fatalf("build account: %v", err)
		}
		id, err := repo.Create(ctx, acc)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if id <= 0 {
			t.Errorf("expected positive id, got %d", id)
		}
	})

	t.Run("duplicate document maps to ErrDuplicateAccount", func(t *testing.T) {
		resetTables(t)
		acc, _ := domain.NewAccount("12345678900")
		if _, err := repo.Create(ctx, acc); err != nil {
			t.Fatalf("seed: %v", err)
		}
		_, err := repo.Create(ctx, acc)
		if !errors.Is(err, domain.ErrDuplicateAccount) {
			t.Fatalf("expected ErrDuplicateAccount, got %v", err)
		}
	})
}

func TestAccountRepository_Integration_FindByID(t *testing.T) {
	repo := NewAccountRepository(testDB)
	ctx := context.Background()

	t.Run("happy path round-trips the account", func(t *testing.T) {
		resetTables(t)
		original, _ := domain.NewAccount("12345678900")
		id, err := repo.Create(ctx, original)
		if err != nil {
			t.Fatalf("seed: %v", err)
		}

		got, err := repo.FindByID(ctx, id)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got.ID != id {
			t.Errorf("ID: want %d, got %d", id, got.ID)
		}
		if got.DocumentNumber != original.DocumentNumber {
			t.Errorf("DocumentNumber: want %q, got %q", original.DocumentNumber, got.DocumentNumber)
		}
		if got.CreatedAt.Location() != time.UTC {
			t.Errorf("CreatedAt not UTC: %v", got.CreatedAt.Location())
		}
		if time.Since(got.CreatedAt) > time.Minute {
			t.Errorf("CreatedAt suspiciously old: %v", got.CreatedAt)
		}
	})

	t.Run("missing id maps to ErrAccountNotFound", func(t *testing.T) {
		resetTables(t)
		_, err := repo.FindByID(ctx, 999999)
		if !errors.Is(err, domain.ErrAccountNotFound) {
			t.Fatalf("expected ErrAccountNotFound, got %v", err)
		}
	})
}
