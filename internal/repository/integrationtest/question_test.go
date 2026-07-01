package integrationtest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/0xpelamar/kingscomp/internal/repository/redis"
	"github.com/ory/dockertest/v4"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestQuestion(t *testing.T) {
	suite.Run(t, new(QuestionTestSuite))
}

type QuestionTestSuite struct {
	suite.Suite
	ctx         context.Context
	redisClient rueidis.Client
	qr          repository.Question
}

func (qts *QuestionTestSuite) SetupSuite() {
	qts.ctx = context.Background()
	pool := dockertest.NewPoolT(qts.T(), "")
	redisResource := pool.RunT(qts.T(), "redis/redis-stack-server", dockertest.WithTag("7.4.0-v8"))

	redisPort := redisResource.GetPort("6379/tcp")

	err := pool.Retry(qts.ctx, 30*time.Second, func() error {
		_, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
		return err
	})
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
	assert.NoError(qts.T(), err)
	qts.redisClient = redisClient
	qts.qr = repository.NewQuestionRedisRepository(redisClient)
}

func (qts *QuestionTestSuite) TestActiveQuestions() {
	count, err := qts.qr.GetActiveQuestionsCount(qts.ctx)
	assert.NoError(qts.T(), err)
	assert.Equal(qts.T(), int64(0), count)

	err = qts.qr.PushActiveQuestion(
		qts.ctx,
		entity.Question{
			ID:            "ID3",
			Question:      "q3",
			Answers:       []string{"ch111", "ch222", "ch333", "ch444"},
			CorrectAnswer: 1,
		},
		entity.Question{
			ID:            "ID2",
			Question:      "q2",
			Answers:       []string{"ch11", "ch22", "ch33", "ch44"},
			CorrectAnswer: 2,
		},
		entity.Question{
			ID:            "ID1",
			Question:      "q1",
			Answers:       []string{"ch1", "ch2", "ch3", "ch4"},
			CorrectAnswer: 3,
		},
	)
	assert.NoError(qts.T(), err)

	count, err = qts.qr.GetActiveQuestionsCount(qts.ctx)
	assert.NoError(qts.T(), err)
	assert.Equal(qts.T(), int64(3), count)

	aqs, err := qts.qr.GetActiveQuestions(qts.ctx, 0, 2)
	assert.NoError(qts.T(), err)
	assert.Len(qts.T(), aqs, 2)
	assert.Equal(qts.T(), "ID3", aqs[0].ID)
	assert.Equal(qts.T(), "ID1", aqs[1].ID)

}
