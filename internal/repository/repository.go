package repository

import (
	"context"
	"errors"

	"github.com/0xpelamar/kingscomp/internal/entity"
)

var ErrNotFound = errors.New("entity not found")

type CommonBehaviour[T entity.Entity] interface {
	Get(context.Context, entity.ID) (T, error)
	Save(context.Context, T) error
}

type AccountRepository interface {
	CommonBehaviour[entity.Account]
}
