package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID              int64
	AccountID       int64
	OperationTypeID OperationType
	Amount          decimal.Decimal // signed: negative for 1/2/3, positive for 4
	EventDate       time.Time
}

// NewTransaction builds a valid Transaction from client input.
// The caller passes a positive amount; the sign is applied here based on
// the operation type — this is the single source of truth for the rule.
func NewTransaction(accountID int64, opType OperationType, amount decimal.Decimal) (Transaction, error) {
	if accountID <= 0 {
		return Transaction{}, fmt.Errorf("%w: must be positive", ErrInvalidAccountID)
	}
	if !opType.Valid() {
		return Transaction{}, fmt.Errorf("%w: got %d", ErrInvalidOperationType, opType)
	}
	if !amount.IsPositive() {
		return Transaction{}, fmt.Errorf("%w: must be greater than zero", ErrInvalidAmount)
	}
	if !amount.Truncate(2).Equal(amount) {
		return Transaction{}, fmt.Errorf("%w: at most 2 decimal places", ErrInvalidAmount)
	}

	signed := amount
	if opType.Sign() == -1 {
		signed = amount.Neg()
	}

	return Transaction{
		AccountID:       accountID,
		OperationTypeID: opType,
		Amount:          signed,
		EventDate:       time.Now().UTC(),
	}, nil
}
