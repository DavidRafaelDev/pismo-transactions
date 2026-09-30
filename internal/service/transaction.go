package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
	"github.com/shopspring/decimal"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx domain.Transaction) (int64, error)
}

type TransactionService struct {
	accounts     AccountRepository
	transactions TransactionRepository
}

func NewTransactionService(accounts AccountRepository, transactions TransactionRepository) *TransactionService {
	return &TransactionService{accounts: accounts, transactions: transactions}
}

func (s *TransactionService) Create(
	ctx context.Context,
	accountID int64,
	opType domain.OperationType,
	amount decimal.Decimal,
) (domain.Transaction, error) {
	// Explicit existence check so "account not found" maps cleanly to 422
	// instead of leaking a foreign-key constraint error from the DB layer.
	if _, err := s.accounts.FindByID(ctx, accountID); err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return domain.Transaction{}, err
		}
		return domain.Transaction{}, fmt.Errorf("transaction.create.check_account: %w", err)
	}

	tx, err := domain.NewTransaction(accountID, opType, amount)
	if err != nil {
		return domain.Transaction{}, err
	}

	id, err := s.transactions.Create(ctx, tx)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("transaction.create: %w", err)
	}
	tx.ID = id
	return tx, nil
}
