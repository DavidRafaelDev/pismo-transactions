package httpapi

import (
	"errors"
	"net/http"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
)

// writeError translates a domain error into an HTTP response.
// Endpoint-specific overrides (like ErrAccountNotFound mapping to 422 on
// POST /transactions instead of 404) are handled by the specific handler
// BEFORE delegating to this generic mapper.
func writeError(w http.ResponseWriter, err error) {
	code, status, message := mapError(err)
	writeJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}

func mapError(err error) (code string, status int, message string) {
	switch {
	case errors.Is(err, domain.ErrInvalidDocumentNumber):
		return "invalid_document_number", http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrInvalidAmount):
		return "invalid_amount", http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrInvalidOperationType):
		return "invalid_operation_type", http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrInvalidAccountID):
		return "invalid_account_id", http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrAccountNotFound):
		return "account_not_found", http.StatusNotFound, "account not found"
	case errors.Is(err, domain.ErrDuplicateAccount):
		return "duplicate_account", http.StatusConflict, "account already exists"
	default:
		return "internal_error", http.StatusInternalServerError, "unexpected error"
	}
}
