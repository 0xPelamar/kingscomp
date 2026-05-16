package integrationtest

import (
	"context"
	"fmt"
	"strconv"
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

type TestType struct {
	ID   string
	Name string
}

func (t TestType) EntityID() entity.ID {
	return entity.NewID("testType", t.ID)
}

func TestCommonBehaviour(t *testing.T) {
	suite.Run(t, new(CommonBehaviourTestSuite))
}

type CommonBehaviourTestSuite struct {
	suite.Suite
	ctx         context.Context
	redisClient rueidis.Client
	rcb         *repository.RedisCommonBehaviour[TestType]
}

func (s *CommonBehaviourTestSuite) SetupSuite() {
	s.ctx = context.Background()
	pool := dockertest.NewPoolT(s.T(), "")
	redisResource := pool.RunT(s.T(), "redis", dockertest.WithTag("8.4-alpine"))

	redisPort := redisResource.GetPort("6379/tcp")

	err := pool.Retry(s.ctx, 30*time.Second, func() error {
		_, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
		return err
	})
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
	assert.NoError(s.T(), err)
	s.redisClient = redisClient
	s.rcb = repository.NewRedisCommonBehaviour[TestType](redisClient)
}
func (s *CommonBehaviourTestSuite) TearDownTest() {
	flushAll(s.T(), s.redisClient)

}
func flushAll(t *testing.T, redisClient rueidis.Client) {
	assert.NoError(t, redisClient.Do(t.Context(), redisClient.B().Flushall().Build()).Error())
}

func (s *CommonBehaviourTestSuite) TestRedisCommonBehaviour_GetAndSave() {

	err := s.rcb.Save(s.ctx, TestType{"21", "asdf"})
	assert.NoError(s.T(), err)

	err = s.rcb.Save(s.ctx, TestType{"22", "qwer"})
	assert.NoError(s.T(), err)

	ins, err := s.rcb.Get(s.ctx, entity.NewID("testType", "21"))
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "asdf", ins.Name)
	assert.Equal(s.T(), "21", ins.ID)

	ins, err = s.rcb.Get(s.ctx, entity.NewID("testType", "22"))
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "qwer", ins.Name)
	assert.Equal(s.T(), "22", ins.ID)

	ins, err = s.rcb.Get(s.ctx, entity.NewID("testType", "23"))
	assert.ErrorIs(s.T(), err, repository.ErrNotFound)

}

func (s *CommonBehaviourTestSuite) TestRedisCommonBehaviour_Mget() {
	var err error
	for i := 0; i < 10; i++ {
		err = s.rcb.Save(s.ctx, TestType{
			ID:   strconv.Itoa(i),
			Name: fmt.Sprintf("name-%d", i),
		})
		assert.NoError(s.T(), err)
	}

	items, err := s.rcb.Mget(s.ctx,
		entity.NewID("testType", 2),
		entity.NewID("testType", 3),
		entity.NewID("testType", 4),
		entity.NewID("testType", 5),
	)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), items, 4)
	assert.Equal(s.T(), "name-3", items[1].Name)
}

func (s *CommonBehaviourTestSuite) TestRedisCommonBehaviour_MgetNotExists() {
	var err error

	items, err := s.rcb.Mget(s.ctx,
		entity.NewID("testType", 2),
		entity.NewID("testType", 3),
		entity.NewID("testType", 4),
		entity.NewID("testType", 5),
	)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), items, 0)
}

func (s *CommonBehaviourTestSuite) TestRedisCommonBehaviour_Mset() {
	err := s.rcb.Mset(s.ctx,
		TestType{ID: "2", Name: "name2"},
		TestType{ID: "3", Name: "name3"},
		TestType{ID: "4", Name: "name4"},
		TestType{ID: "5", Name: "name5"},
	)
	assert.NoError(s.T(), err)

	entities, err := s.rcb.Mget(s.ctx,
		entity.NewID("testType", 2),
		entity.NewID("testType", 3),
		entity.NewID("testType", 4),
		entity.NewID("testType", 5))
	assert.NoError(s.T(), err)
	assert.Len(s.T(), entities, 4)
	assert.Equal(s.T(), "2", entities[0].ID)
	assert.Equal(s.T(), "name2", entities[0].Name)
}
