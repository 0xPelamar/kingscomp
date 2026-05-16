package repository

import (
	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/redis/rueidis"
)

var _ Question = &QuestionRedisRepository{}

type QuestionRedisRepository struct {
	*RedisCommonBehaviour[entity.Question]
}

func NewQuestionRedisRepository(client rueidis.Client) *QuestionRedisRepository {
	return &QuestionRedisRepository{
		NewRedisCommonBehaviour[entity.Question](client),
	}
}
