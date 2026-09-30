//go:build integration

package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
	"github.com/shopspring/decimal"
)

func TestTransactionRepository_Integration_Create(t *testing.T) {
	resetTables(t)
	accRepo := NewAccountRepository(testDB)
	txRepo := NewTransactionRepository(testDB)
	ctx := context.Background()

	seedAcc, _ := domain.NewAccount("12345678900")
	accountID, err := accRepo.Create(ctx, seedAcc)
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}

	mustDecimal := func(s string) decimal.Decimal {
		d, err := decimal.NewFromString(s)
		if err != nil {
			t.Fatalf("bad decimal %q: %v", s, err)
		}
		return d
	}

	tests := []struct {
		name         string
		opType       domain.OperationType
		clientAmount string
		wantSigned   string // expected value in DB after sign is applied
	}{
		{"normal purchase persists negative amount", domain.NormalPurchase, "50.00", "-50.00"},
		{"purchase with installments persists negative amount", domain.PurchaseWithInstallments, "23.50", "-23.50"},
		{"withdrawal persists negative amount", domain.Withdrawal, "18.70", "-18.70"},
		{"credit voucher persists positive amount", domain.CreditVoucher, "60.00", "60.00"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := domain.NewTransaction(accountID, tc.opType, mustDecimal(tc.clientAmount))
			if err != nil {
				t.Fatalf("build transaction: %v", err)
			}

			id, err := txRepo.Create(ctx, tx)
			if err != nil {
				t.Fatalf("Create: %v", err)
			}
			if id <= 0 {
				t.Fatalf("expected positive id, got %d", id)
			}

			// Round-trip: read the row back and verify what is stored.
			var (
				dbAccountID int64
				dbOpType    int8
				dbAmount    decimal.Decimal
				dbEventDate time.Time
			)
			err = testDB.QueryRowContext(ctx,
				`SELECT account_id, operation_type_id, amount, event_date
				 FROM transactions WHERE transaction_id = ?`, id,
			).Scan(&dbAccountID, &dbOpType, &dbAmount, &dbEventDate)
			if err != nil {
				t.Fatalf("read back: %v", err)
			}

			if dbAccountID != accountID {
				t.Errorf("account_id: want %d, got %d", accountID, dbAccountID)
			}
			if dbOpType != int8(tc.opType) {
				t.Errorf("operation_type_id: want %d, got %d", tc.opType, dbOpType)
			}
			if !dbAmount.Equal(mustDecimal(tc.wantSigned)) {
				t.Errorf("amount: want %s, got %s", tc.wantSigned, dbAmount)
			}
			if dbEventDate.Location() != time.UTC {
				t.Errorf("event_date not UTC: %v", dbEventDate.Location())
			}
			// MySQL DATETIME(6) is microsecond precision; Go time.Now() is
			// nanosecond precision. Small drift is expected. 1s window is
			// paranoid enough for a real-world write/read cycle.
			if diff := dbEventDate.Sub(tx.EventDate).Abs(); diff > time.Second {
				t.Errorf("event_date drift too large: db=%s, in=%s (diff=%s)",
					dbEventDate, tx.EventDate, diff)
			}
		})
	}
}
