package service

import (
	"context"
	"testing"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountService_CreateOrUpdateWithUserExists(t *testing.T) {
	accRep := repository.NewMockAccount(t)
	s := NewAccountService(accRep)
	ctx := context.Background()
	accRep.On("Get", mock.Anything, entity.NewID("account", 11)).
		Return(entity.Account{ID: 11, FirstName: "whatever"}, nil).Once()
	accRep.On("Save", mock.Anything, mock.MatchedBy(func(acc entity.Account) bool {
		return acc.FirstName == "changed"
	})).Return(nil).Once()

	newAcc, created, err := s.CreateOrUpdate(ctx, entity.Account{
		ID:        11,
		FirstName: "changed",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(11), newAcc.ID)
	assert.Equal(t, "changed", newAcc.FirstName)
	assert.False(t, created)

	accRep.AssertExpectations(t)
}

func TestAccountService_CreateOrUpdateWithUserNotExists(t *testing.T) {
	accRep := repository.NewMockAccount(t)
	s := NewAccountService(accRep)
	ctx := context.Background()
	accRep.On("Get", mock.Anything, entity.NewID("account", 11)).
		Return(entity.Account{}, repository.ErrNotFound).Once()
	accRep.On("Save", mock.Anything, mock.MatchedBy(func(acc entity.Account) bool {
		return acc.FirstName == "newacc"
	})).Return(nil).Once()

	newAcc, created, err := s.CreateOrUpdate(ctx, entity.Account{
		ID:        11,
		FirstName: "newacc",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(11), newAcc.ID)
	assert.Equal(t, "newacc", newAcc.FirstName)
	assert.True(t, created)

	accRep.AssertExpectations(t)
}

func TestAccountService_CreateOrUpdateUserHasNotChanged(t *testing.T) {
	accRep := repository.NewMockAccount(t)
	s := NewAccountService(accRep)
	ctx := context.Background()
	accRep.On("Get", mock.Anything, entity.NewID("account", 11)).
		Return(entity.Account{ID: 11, FirstName: "whatever"}, nil).Once()

	newAcc, created, err := s.CreateOrUpdate(ctx, entity.Account{
		ID:        11,
		FirstName: "whatever",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(11), newAcc.ID)
	assert.Equal(t, "whatever", newAcc.FirstName)
	assert.False(t, created)

	accRep.AssertExpectations(t)
}
