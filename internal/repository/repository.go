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
	MGet(ctx context.Context, IDs ...entity.ID) ([]T, error)
	MSet(ctx context.Context, entities ...T) error
}

type Account interface {
	CommonBehaviour[entity.Account]
}

type Lobby interface {
	CommonBehaviour[entity.Lobby]
	LobbyPlayers(ctx context.Context, LobbyID entity.ID) ([]entity.Account, error)
}

type Question interface {
	CommonBehaviour[entity.Question]
	GetActiveQuestionsCount(ctx context.Context) (int64, error)
	GetActiveQuestions(ctx context.Context, index ...int64) ([]entity.Question, error)
	PushActiveQuestion(ctx context.Context, questions ...entity.Question) error
}
