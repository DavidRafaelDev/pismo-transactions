package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Up applies all pending "up" migrations found under the given filesystem
// at the "migrations" subdirectory. Idempotent: returns nil when already up to date.
func Up(database *sql.DB, files fs.FS) error {
	src, err := iofs.New(files, "migrations")
	if err != nil {
		return fmt.Errorf("iofs.New: %w", err)
	}

	driver, err := mysqlmigrate.WithInstance(database, &mysqlmigrate.Config{})
	if err != nil {
		return fmt.Errorf("mysql.WithInstance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithInstance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
