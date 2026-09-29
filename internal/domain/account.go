package domain

import (
	"fmt"
	"regexp"
	"time"
)

// documentPattern accepts only digits, with length 11 (CPF) or 14 (CNPJ).
var documentPattern = regexp.MustCompile(`^(?:\d{11}|\d{14})$`)

type Account struct {
	ID             int64
	DocumentNumber string
	CreatedAt      time.Time
}

// NewAccount builds a valid Account from an untrusted document number.
// ID is zero until persisted; CreatedAt is set to now in UTC.
func NewAccount(documentNumber string) (Account, error) {
	if !documentPattern.MatchString(documentNumber) {
		return Account{}, fmt.Errorf("%w: expected 11 or 14 digits", ErrInvalidDocumentNumber)
	}
	return Account{
		DocumentNumber: documentNumber,
		CreatedAt:      time.Now().UTC(),
	}, nil
}
