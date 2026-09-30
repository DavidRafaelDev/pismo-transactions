package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx domain.Transaction) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO transactions (account_id, operation_type_id, amount, event_date)
		 VALUES (?, ?, ?, ?)`,
		tx.AccountID,
		int8(tx.OperationTypeID),
		tx.Amount, // shopspring/decimal implements driver.Valuer (writes as string)
		tx.EventDate,
	)
	if err != nil {
		return 0, fmt.Errorf("insert transaction: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read last insert id: %w", err)
	}
	return id, nil
}
