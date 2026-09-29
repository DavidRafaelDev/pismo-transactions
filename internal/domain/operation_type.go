package domain

// OperationType identifies the kind of a transaction.
// Values are seeded in the operation_types table by migration 0003.
type OperationType int8

const (
	NormalPurchase           OperationType = 1
	PurchaseWithInstallments OperationType = 2
	Withdrawal               OperationType = 3
	CreditVoucher            OperationType = 4
)

func (o OperationType) Valid() bool {
	return o >= NormalPurchase && o <= CreditVoucher
}

// Sign returns +1 for credit vouchers and -1 for the other three types.
// This is the single source of truth for the transaction sign rule.
func (o OperationType) Sign() int {
	if o == CreditVoucher {
		return 1
	}
	return -1
}
