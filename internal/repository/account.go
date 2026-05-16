package repository

import (
	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/redis/rueidis"
)

var _ Account = &AccountRedisRepository{}

type AccountRedisRepository struct {
	*RedisCommonBehaviour[entity.Account]
}

func NewAccountRedisRepository(client rueidis.Client) *AccountRedisRepository {
	return &AccountRedisRepository{
		NewRedisCommonBehaviour[entity.Account](client),
	}
}
