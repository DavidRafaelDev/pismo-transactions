package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
)

const maxBodyBytes = 1 << 20 // 1 MiB

// accountService is defined here (consumer-side) so this package does not
// depend on the concrete service implementation.
type accountService interface {
	Create(ctx context.Context, documentNumber string) (domain.Account, error)
	Get(ctx context.Context, id int64) (domain.Account, error)
}

type AccountsHandler struct {
	svc accountService
}

func NewAccountsHandler(svc accountService) *AccountsHandler {
	return &AccountsHandler{svc: svc}
}

func (h *AccountsHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req createAccountRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: errorBody{
			Code:    "invalid_body",
			Message: err.Error(),
		}})
		return
	}

	acc, err := h.svc.Create(r.Context(), req.DocumentNumber)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, accountResponse{
		AccountID:      acc.ID,
		DocumentNumber: acc.DocumentNumber,
	})
}

func (h *AccountsHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: errorBody{
			Code:    "invalid_account_id",
			Message: "id must be a positive integer",
		}})
		return
	}
	if id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: errorBody{
			Code:    "invalid_account_id",
			Message: "id must be a positive integer",
		}})
		return
	}

	acc, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, accountResponse{
		AccountID:      acc.ID,
		DocumentNumber: acc.DocumentNumber,
	})
}
