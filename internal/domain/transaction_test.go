package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func mustDecimal(t *testing.T, s string) decimal.Decimal {
	t.Helper()
	d, err := decimal.NewFromString(s)
	if err != nil {
		t.Fatalf("bad decimal literal %q: %v", s, err)
	}
	return d
}

func TestNewTransaction_ValidatesInputs(t *testing.T) {
	tests := []struct {
		name      string
		accountID int64
		opType    OperationType
		amount    string
		wantErr   error
	}{
		{"account id zero", 0, NormalPurchase, "10.00", ErrInvalidAccountID},
		{"account id negative", -1, NormalPurchase, "10.00", ErrInvalidAccountID},
		{"op type zero", 1, 0, "10.00", ErrInvalidOperationType},
		{"op type out of range", 1, 5, "10.00", ErrInvalidOperationType},
		{"amount zero", 1, NormalPurchase, "0", ErrInvalidAmount},
		{"amount negative", 1, NormalPurchase, "-10.00", ErrInvalidAmount},
		{"amount 3 decimals", 1, NormalPurchase, "10.001", ErrInvalidAmount},
		{"amount 4 decimals", 1, NormalPurchase, "10.1234", ErrInvalidAmount},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewTransaction(tc.accountID, tc.opType, mustDecimal(t, tc.amount))
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err: want %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestNewTransaction_AppliesSign(t *testing.T) {
	tests := []struct {
		name         string
		opType       OperationType
		input        string
		signedAmount string
	}{
		{"normal purchase -> negative", NormalPurchase, "50.00", "-50.00"},
		{"installments -> negative", PurchaseWithInstallments, "23.50", "-23.50"},
		{"withdrawal -> negative", Withdrawal, "18.70", "-18.70"},
		{"credit voucher -> positive", CreditVoucher, "60.00", "60.00"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := NewTransaction(1, tc.opType, mustDecimal(t, tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := mustDecimal(t, tc.signedAmount)
			if !tx.Amount.Equal(want) {
				t.Errorf("Amount: want %s, got %s", want, tx.Amount)
			}
			if tx.AccountID != 1 {
				t.Errorf("AccountID: want 1, got %d", tx.AccountID)
			}
			if tx.OperationTypeID != tc.opType {
				t.Errorf("OperationTypeID: want %d, got %d", tc.opType, tx.OperationTypeID)
			}
			if time.Since(tx.EventDate) > time.Second {
				t.Errorf("EventDate too old: %v", tx.EventDate)
			}
			if tx.EventDate.Location() != time.UTC {
				t.Errorf("EventDate not UTC: %v", tx.EventDate.Location())
			}
			if tx.ID != 0 {
				t.Errorf("ID should be zero before persist, got %d", tx.ID)
			}
		})
	}
}
