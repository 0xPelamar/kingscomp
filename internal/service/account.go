package service

import (
	"context"
	"errors"
	"time"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/repository"
)

const (
	DefaultState = "home"
)

type AccountService struct {
	accounts repository.AccountRepository
}

func NewAccountService(accountRepository repository.AccountRepository) *AccountService {
	return &AccountService{accounts: accountRepository}

}

func (a *AccountService) CreateOrUpdate(ctx context.Context, account entity.Account) (entity.Account, bool, error) {
	savedAccount, err := a.accounts.Get(ctx, account.EntityID())
	// user exists in the database
	if err == nil {
		if savedAccount.FirstName != account.FirstName ||
			savedAccount.LastName != account.LastName ||
			savedAccount.Username != account.Username {
			savedAccount.Username = account.Username
			savedAccount.FirstName = account.FirstName
			savedAccount.LastName = account.LastName
			return savedAccount, false, a.accounts.Save(ctx, savedAccount)
		}
		return savedAccount, false, nil
	}

	// user does not exist in the database
	if errors.Is(err, repository.ErrNotFound) {
		account.JoinedAt = time.Now()
		account.State = DefaultState
		return account, true, a.accounts.Save(ctx, account)
	}
	return entity.Account{}, false, err
}

func (a *AccountService) Update(ctx context.Context, account entity.Account) error {
	return a.accounts.Save(ctx, account)
}
