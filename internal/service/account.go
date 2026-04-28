package service

import (
	"context"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/repository"
)

type AccountService struct {
	accounts repository.AccountRepository
}

func NewAccountService(accountRepository repository.AccountRepository) *AccountService {
	return &AccountService{accounts: accountRepository}

}

func (a *AccountService) CreateOrUpdate(ctx context.Context, account entity.Account) error {
	return a.accounts.Save(ctx, account)
}
