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

func (cbts *CommonBehaviourTestSuite) SetupSuite() {
	cbts.ctx = context.Background()
	pool := dockertest.NewPoolT(cbts.T(), "")
	redisResource := pool.RunT(cbts.T(), "redis/redis-stack-server", dockertest.WithTag("7.4.0-v8"))

	redisPort := redisResource.GetPort("6379/tcp")

	err := pool.Retry(cbts.ctx, 30*time.Second, func() error {
		_, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
		return err
	})
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
	assert.NoError(cbts.T(), err)
	cbts.redisClient = redisClient
	cbts.rcb = repository.NewRedisCommonBehaviour[TestType](redisClient)
}

func (cbts *CommonBehaviourTestSuite) TearDownTest() {
	flushAll(cbts.T(), cbts.redisClient)
}

func (cbts *CommonBehaviourTestSuite) TestRedisCommonBehaviour_GetAndSave() {

	err := cbts.rcb.Save(cbts.ctx, TestType{"21", "asdf"})
	assert.NoError(cbts.T(), err)

	err = cbts.rcb.Save(cbts.ctx, TestType{"22", "qwer"})
	assert.NoError(cbts.T(), err)

	ins, err := cbts.rcb.Get(cbts.ctx, entity.NewID("testType", "21"))
	assert.NoError(cbts.T(), err)
	assert.Equal(cbts.T(), "asdf", ins.Name)
	assert.Equal(cbts.T(), "21", ins.ID)

	ins, err = cbts.rcb.Get(cbts.ctx, entity.NewID("testType", "22"))
	assert.NoError(cbts.T(), err)
	assert.Equal(cbts.T(), "qwer", ins.Name)
	assert.Equal(cbts.T(), "22", ins.ID)

	ins, err = cbts.rcb.Get(cbts.ctx, entity.NewID("testType", "23"))
	assert.ErrorIs(cbts.T(), err, repository.ErrNotFound)

}

func (cbts *CommonBehaviourTestSuite) TestRedisCommonBehaviour_Mget() {
	var err error
	for i := 0; i < 10; i++ {
		err = cbts.rcb.Save(cbts.ctx, TestType{
			ID:   strconv.Itoa(i),
			Name: fmt.Sprintf("name-%d", i),
		})
		assert.NoError(cbts.T(), err)
	}

	items, err := cbts.rcb.MGet(cbts.ctx,
		entity.NewID("testType", 2),
		entity.NewID("testType", 3),
		entity.NewID("testType", 4),
		entity.NewID("testType", 5),
	)
	assert.NoError(cbts.T(), err)
	assert.Len(cbts.T(), items, 4)
	assert.Equal(cbts.T(), "name-3", items[1].Name)
}

func (cbts *CommonBehaviourTestSuite) TestRedisCommonBehaviour_MgetNotExists() {
	var err error

	items, err := cbts.rcb.MGet(cbts.ctx,
		entity.NewID("testType", 2),
		entity.NewID("testType", 3),
		entity.NewID("testType", 4),
		entity.NewID("testType", 5),
	)
	assert.NoError(cbts.T(), err)
	assert.Len(cbts.T(), items, 0)
}

func (cbts *CommonBehaviourTestSuite) TestRedisCommonBehaviour_Mset() {
	err := cbts.rcb.MSet(cbts.ctx,
		TestType{ID: "2", Name: "name2"},
		TestType{ID: "3", Name: "name3"},
		TestType{ID: "4", Name: "name4"},
		TestType{ID: "5", Name: "name5"},
	)
	assert.NoError(cbts.T(), err)

	entities, err := cbts.rcb.MGet(cbts.ctx,
		entity.NewID("testType", 2),
		entity.NewID("testType", 3),
		entity.NewID("testType", 4),
		entity.NewID("testType", 5))
	assert.NoError(cbts.T(), err)
	assert.Len(cbts.T(), entities, 4)
	assert.Equal(cbts.T(), "2", entities[0].ID)
	assert.Equal(cbts.T(), "name2", entities[0].Name)
}

func (cbts *CommonBehaviourTestSuite) TestRedisCommonBehaviour_SetField() {
	err := cbts.rcb.MSet(cbts.ctx,
		TestType{ID: "2", Name: "name2"},
		TestType{ID: "3", Name: "name3"},
		TestType{ID: "4", Name: "name4"},
		TestType{ID: "5", Name: "name5"},
	)
	assert.NoError(cbts.T(), err)

	err = cbts.rcb.SetField(cbts.ctx, entity.NewID("testType", "2"), "Name", "updatedName")
	assert.NoError(cbts.T(), err)
	acc2, err := cbts.rcb.Get(cbts.ctx, entity.NewID("testType", "2"))
	assert.NoError(cbts.T(), err)
	assert.Equal(cbts.T(), "updatedName", acc2.Name)
}
