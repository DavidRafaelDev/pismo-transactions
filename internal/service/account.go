package service

import (
	"context"
	"fmt"

	"github.com/DavidRafaelDev/pismo-transactions/internal/domain"
)

// AccountRepository is defined here (the consumer) so the service package
// stays decoupled from any specific persistence implementation.
type AccountRepository interface {
	Create(ctx context.Context, acc domain.Account) (int64, error)
	FindByID(ctx context.Context, id int64) (domain.Account, error)
}

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) Create(ctx context.Context, documentNumber string) (domain.Account, error) {
	acc, err := domain.NewAccount(documentNumber)
	if err != nil {
		return domain.Account{}, err
	}
	id, err := s.repo.Create(ctx, acc)
	if err != nil {
		return domain.Account{}, fmt.Errorf("account.create: %w", err)
	}
	acc.ID = id
	return acc, nil
}

func (s *AccountService) Get(ctx context.Context, id int64) (domain.Account, error) {
	if id <= 0 {
		return domain.Account{}, fmt.Errorf("%w: must be positive", domain.ErrInvalidAccountID)
	}
	acc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Account{}, fmt.Errorf("account.get: %w", err)
	}
	return acc, nil
}
