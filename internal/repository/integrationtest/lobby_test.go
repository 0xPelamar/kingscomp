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

func TestLobby(t *testing.T) {
	suite.Run(t, new(LobbyTestSuite))
}

type LobbyTestSuite struct {
	suite.Suite
	ctx         context.Context
	redisClient rueidis.Client
	lr          repository.Lobby
}

func (lts *LobbyTestSuite) SetupSuite() {
	lts.ctx = context.Background()
	pool := dockertest.NewPoolT(lts.T(), "")
	redisResource := pool.RunT(lts.T(), "redis/redis-stack-server", dockertest.WithTag("7.4.0-v8"))

	redisPort := redisResource.GetPort("6379/tcp")

	err := pool.Retry(lts.ctx, 30*time.Second, func() error {
		_, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
		return err
	})
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
	assert.NoError(lts.T(), err)
	lts.redisClient = redisClient
	lts.lr = repository.NewLobbyRedisRepository(redisClient)
}

func (lts *LobbyTestSuite) TestLobby_Ready() {
	err := lts.lr.Save(lts.ctx, entity.Lobby{
		ID:           "1",
		Participants: []int64{1, 2},
		UserState: map[int64]entity.UserState{
			1: entity.UserState{},
			2: entity.UserState{},
		},
	})
	assert.NoError(lts.T(), err)

	err = lts.lr.UpdateUserState(lts.ctx, "1", 1, "is_ready", true)
	assert.NoError(lts.T(), err)

	lobby, err := lts.lr.Get(lts.ctx, entity.NewID("lobby", 1))
	assert.NoError(lts.T(), err)
	assert.Equal(lts.T(), true, lobby.UserState[1].IsReady)
}
