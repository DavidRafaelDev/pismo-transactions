package domain

import "testing"

func TestOperationType_Valid(t *testing.T) {
	tests := []struct {
		name string
		op   OperationType
		want bool
	}{
		{"normal purchase", NormalPurchase, true},
		{"purchase with installments", PurchaseWithInstallments, true},
		{"withdrawal", Withdrawal, true},
		{"credit voucher", CreditVoucher, true},
		{"zero is invalid", 0, false},
		{"five is invalid", 5, false},
		{"negative is invalid", -1, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.op.Valid(); got != tc.want {
				t.Errorf("Valid() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOperationType_Sign(t *testing.T) {
	tests := []struct {
		name string
		op   OperationType
		want int
	}{
		{"normal purchase is negative", NormalPurchase, -1},
		{"installments is negative", PurchaseWithInstallments, -1},
		{"withdrawal is negative", Withdrawal, -1},
		{"credit voucher is positive", CreditVoucher, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.op.Sign(); got != tc.want {
				t.Errorf("Sign() = %d, want %d", got, tc.want)
			}
		})
	}
}
