package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
	sqldriver "github.com/go-sql-driver/mysql"
)

// duplicateEntryCode is the MySQL error number for a violation of a
// UNIQUE constraint (ER_DUP_ENTRY). We use it to translate DB-level
// errors into domain.ErrDuplicateAccount without leaking driver types.
const duplicateEntryCode = 1062

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, acc domain.Account) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts (document_number, created_at) VALUES (?, ?)`,
		acc.DocumentNumber, acc.CreatedAt,
	)
	if err != nil {
		var mysqlErr *sqldriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == duplicateEntryCode {
			return 0, domain.ErrDuplicateAccount
		}
		return 0, fmt.Errorf("insert account: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read last insert id: %w", err)
	}
	return id, nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id int64) (domain.Account, error) {
	var acc domain.Account
	err := r.db.QueryRowContext(ctx,
		`SELECT account_id, document_number, created_at FROM accounts WHERE account_id = ?`,
		id,
	).Scan(&acc.ID, &acc.DocumentNumber, &acc.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("select account: %w", err)
	}
	return acc, nil
}
