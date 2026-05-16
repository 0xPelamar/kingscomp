package repository

import (
	"context"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/redis/rueidis"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
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

func (q QuestionRedisRepository) GetActiveQuestionsCount(ctx context.Context) (int64, error) {
	cmd := q.client.B().Llen().Key("active_questions").Build()
	return q.client.Do(ctx, cmd).ToInt64()
}

func (q QuestionRedisRepository) GetActiveQuestions(ctx context.Context, indexes ...int64) ([]entity.Question, error) {
	cmds := lo.Map(indexes, func(id int64, _ int) rueidis.Completed {
		return q.client.B().Lindex().Key("active_questions").Index(id).Build()
	})
	resp := q.client.DoMulti(ctx, cmds...)
	err := lo.Reduce(resp, func(agg error, item rueidis.RedisResult, _ int) error {
		if agg != nil {
			return agg
		}
		return item.Error()
	}, nil)
	if err != nil {
		logrus.WithError(err).Errorln("failed to load active questions")
		return nil, err
	}

	questionIDs := lo.Map(resp, func(item rueidis.RedisResult, _ int) entity.ID {
		s, _ := item.ToString()
		return entity.NewID("question", s)
	})
	return q.Mget(ctx, questionIDs...)

}

func (q QuestionRedisRepository) PushActiveQuestion(ctx context.Context, questions ...entity.Question) error {
	panic("implement me")
}
