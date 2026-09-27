package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/DavidRafaelDev/pismo-transactions/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	addr := envOr("HTTP_ADDR", ":8080")

	srv := &http.Server{
		Addr:    addr,
		Handler: httpapi.NewRouter(),
	}

	logger.Info("api starting", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server exited", "err", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
