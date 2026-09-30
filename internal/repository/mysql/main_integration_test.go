//go:build integration

package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/DavidRafaelDev/pismo-transactions/internal/config"
	"github.com/DavidRafaelDev/pismo-transactions/internal/db"
)

var testDB *sql.DB

// TestMain runs once per package. It opens the shared DB connection used
// by every integration test and closes it at the end. Failing here means
// the compose stack is not up.
func TestMain(m *testing.M) {
	cfg := config.Load()
	ctx := context.Background()

	conn, err := db.Open(ctx, cfg.DSN())
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"integration tests require MySQL up (`make up`): %v\n", err)
		os.Exit(1)
	}
	testDB = conn

	code := m.Run()

	_ = conn.Close()
	os.Exit(code)
}

// resetTables clears all mutable data between tests.
// Order matters because of the FK from transactions -> accounts.
func resetTables(t *testing.T) {
	t.Helper()
	if _, err := testDB.Exec("DELETE FROM transactions"); err != nil {
		t.Fatalf("reset transactions: %v", err)
	}
	if _, err := testDB.Exec("DELETE FROM accounts"); err != nil {
		t.Fatalf("reset accounts: %v", err)
	}
}
