package httpapi

import (
	"log/slog"
	"net/http"
)

func NewRouter(
	p Pinger,
	accounts *AccountsHandler,
	transactions *TransactionsHandler,
	logger *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", HealthHandler(p))
	mux.HandleFunc("POST /accounts", accounts.Create)
	mux.HandleFunc("GET /accounts/{id}", accounts.Get)
	mux.HandleFunc("POST /transactions", transactions.Create)
	return withLogging(mux, logger)
}
