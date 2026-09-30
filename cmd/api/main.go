package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	pismotx "github.com/DavidRafaelDev/pismo-transactions"
	"github.com/DavidRafaelDev/pismo-transactions/internal/config"
	"github.com/DavidRafaelDev/pismo-transactions/internal/db"
	"github.com/DavidRafaelDev/pismo-transactions/internal/httpapi"
	"github.com/DavidRafaelDev/pismo-transactions/internal/migrate"
	mysqlrepo "github.com/DavidRafaelDev/pismo-transactions/internal/repository/mysql"
	"github.com/DavidRafaelDev/pismo-transactions/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Load()
	ctx := context.Background()

	database, err := db.Open(ctx, cfg.DSN())
	if err != nil {
		logger.Error("db.open", "err", err)
		os.Exit(1)
	}
	defer database.Close()
	logger.Info("db.connected", "host", cfg.DBHost, "name", cfg.DBName)

	if err := migrate.Up(database, pismotx.Migrations); err != nil {
		logger.Error("migrations", "err", err)
		os.Exit(1)
	}
	logger.Info("migrations.applied")

	// Wire: repositories -> services -> handlers -> router.
	accountRepo := mysqlrepo.NewAccountRepository(database)
	transactionRepo := mysqlrepo.NewTransactionRepository(database)

	accountSvc := service.NewAccountService(accountRepo)
	transactionSvc := service.NewTransactionService(accountRepo, transactionRepo)

	accountsHandler := httpapi.NewAccountsHandler(accountSvc)
	transactionsHandler := httpapi.NewTransactionsHandler(transactionSvc)

	router := httpapi.NewRouter(database, accountsHandler, transactionsHandler, logger)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	logger.Info("api.starting", "addr", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server.exited", "err", err)
		os.Exit(1)
	}
}
