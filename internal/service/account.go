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
	Account repository.AccountRepository
}

func NewAccountService(accountRepository repository.AccountRepository) *AccountService {
	return &AccountService{Account: accountRepository}

}

func (a *AccountService) CreateOrUpdate(ctx context.Context, account entity.Account) (entity.Account, bool, error) {
	savedAccount, err := a.Account.Get(ctx, account.EntityID())
	// user exists in the redis
	if err == nil {
		if savedAccount.FirstName != account.FirstName ||
			savedAccount.LastName != account.LastName ||
			savedAccount.Username != account.Username {
			savedAccount.Username = account.Username
			savedAccount.FirstName = account.FirstName
			savedAccount.LastName = account.LastName
			return savedAccount, false, a.Account.Save(ctx, savedAccount)
		}
		return savedAccount, false, nil
	}

	// user does not exist in the redis
	if errors.Is(err, repository.ErrNotFound) {
		account.JoinedAt = time.Now()
		account.State = DefaultState
		return account, true, a.Account.Save(ctx, account)
	}
	return entity.Account{}, false, err
}

func (a *AccountService) Update(ctx context.Context, account entity.Account) error {
	return a.Account.Save(ctx, account)
}
