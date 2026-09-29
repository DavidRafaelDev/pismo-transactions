package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Open dials MySQL and returns a ready-to-use *sql.DB.
// It retries Ping for a short window to tolerate MySQL cold start
// (docker healthcheck already covers `make up`, but `make run` may race).
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	database, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open mysql: %w", err)
	}

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(5 * time.Minute)

	if err := waitForPing(ctx, database); err != nil {
		_ = database.Close()
		return nil, err
	}
	return database, nil
}

func waitForPing(ctx context.Context, database *sql.DB) error {
	const attempts = 10
	const delay = time.Second

	var lastErr error
	for i := 0; i < attempts; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := database.PingContext(pingCtx)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return fmt.Errorf("database not reachable after %d attempts: %w", attempts, lastErr)
}
