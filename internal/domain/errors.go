package domain

import "errors"

var (
	ErrInvalidDocumentNumber = errors.New("invalid document number")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrInvalidOperationType  = errors.New("invalid operation type")
	ErrInvalidAccountID      = errors.New("invalid account id")
	ErrAccountNotFound       = errors.New("account not found")
	ErrDuplicateAccount      = errors.New("account already exists")
)
