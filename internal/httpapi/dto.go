package httpapi

import (
	"time"

	"github.com/shopspring/decimal"
)

// --- Requests ---

type createAccountRequest struct {
	DocumentNumber string `json:"document_number"`
}

type createTransactionRequest struct {
	AccountID       int64           `json:"account_id"`
	OperationTypeID int8            `json:"operation_type_id"`
	Amount          decimal.Decimal `json:"amount"`
}

// --- Responses ---

type accountResponse struct {
	AccountID      int64  `json:"account_id"`
	DocumentNumber string `json:"document_number"`
}

type transactionResponse struct {
	TransactionID   int64           `json:"transaction_id"`
	AccountID       int64           `json:"account_id"`
	OperationTypeID int8            `json:"operation_type_id"`
	Amount          decimal.Decimal `json:"amount"`
	EventDate       time.Time       `json:"event_date"`
}

// --- Error envelope ---
// Every error response has the shape {"error":{"code":"...","message":"..."}}.

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}
