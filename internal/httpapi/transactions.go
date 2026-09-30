package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
	"github.com/shopspring/decimal"
)

type transactionService interface {
	Create(
		ctx context.Context,
		accountID int64,
		opType domain.OperationType,
		amount decimal.Decimal,
	) (domain.Transaction, error)
}

type TransactionsHandler struct {
	svc transactionService
}

func NewTransactionsHandler(svc transactionService) *TransactionsHandler {
	return &TransactionsHandler{svc: svc}
}

func (h *TransactionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req createTransactionRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: errorBody{
			Code:    "invalid_body",
			Message: err.Error(),
		}})
		return
	}

	tx, err := h.svc.Create(
		r.Context(),
		req.AccountID,
		domain.OperationType(req.OperationTypeID),
		req.Amount,
	)
	if err != nil {
		// For POST /transactions, a missing account is a client input
		// problem (unprocessable entity), not a resource-not-found (404).
		if errors.Is(err, domain.ErrAccountNotFound) {
			writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: errorBody{
				Code:    "account_not_found",
				Message: "account does not exist",
			}})
			return
		}
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, transactionResponse{
		TransactionID:   tx.ID,
		AccountID:       tx.AccountID,
		OperationTypeID: int8(tx.OperationTypeID),
		Amount:          tx.Amount,
		EventDate:       tx.EventDate,
	})
}
