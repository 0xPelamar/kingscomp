package integrationtest

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/matchmaking"
	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/0xpelamar/kingscomp/internal/repository/redis"
	"github.com/ory/dockertest/v4"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const maxLobbySize = 5

func TestMatchmaking(t *testing.T) {
	suite.Run(t, new(MatchmakingTestSuite))
}

type MatchmakingTestSuite struct {
	suite.Suite
	mm          matchmaking.MatchMaking
	ctx         context.Context
	timeout     time.Duration
	redisClient rueidis.Client

	lobby   repository.LobbyRepository
	account repository.AccountRepository
}

func (s *MatchmakingTestSuite) SetupSuite() {
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

	s.timeout = time.Second * 10
	s.redisClient = redisClient
	ar := repository.NewAccountRedisRepository(redisClient)
	lr := repository.NewLobbyRedisRepository(redisClient)
	mm := matchmaking.NewRedisMatchMaking(redisClient, lr)

	for i := 0; i < 100; i++ {
		err := ar.Save(context.Background(), entity.Account{
			ID:        int64(i),
			FirstName: fmt.Sprintf("Name-%d", i),
		})
		assert.NoError(s.T(), err)
	}
	s.mm = mm
	s.lobby = lr
	s.account = ar
}
func (s *MatchmakingTestSuite) TearDownSuite() {
	flushAll(s.T(), s.redisClient)
}

func flushAll(t *testing.T, redisClient rueidis.Client) {
	assert.NoError(t, redisClient.Do(t.Context(), redisClient.B().Flushall().Build()).Error())
}

func (s *MatchmakingTestSuite) TestMatchMaking_Join() {

	var wg sync.WaitGroup
	testJoin := func(id int64) {
		wg.Add(1)
		go func() {
			lobby, _, err := s.mm.Join(s.ctx, id, time.Second)
			assert.NoError(s.T(), err)
			assert.NotEqual(s.T(), "", lobby.ID)
			wg.Done()
		}()
	}
	for i := 0; i < maxLobbySize-1; i++ {
		testJoin(int64(3 + i))
	}
	<-time.After(time.Millisecond * 500)

	assert.Equal(s.T(), int64(maxLobbySize-1), zCount(s.T(), s.redisClient, "matchmaking"))

	lobby, _, err := s.mm.Join(s.ctx, 14, s.timeout)
	assert.NoError(s.T(), err)
	assert.NotEqual(s.T(), "", lobby.ID)
	wg.Wait()
}

func (s *MatchmakingTestSuite) TestMatchmaking_JoinTimeout() {

	var wg sync.WaitGroup
	testJoin := func(id int64) {
		wg.Add(1)
		go func() {
			lobby, _, err := s.mm.Join(s.ctx, id, time.Millisecond*100)
			assert.ErrorIs(s.T(), err, matchmaking.ErrTimeout)
			assert.Equal(s.T(), "", lobby.ID)
			wg.Done()
		}()
	}
	testJoin(10)
	<-time.After(500 * time.Millisecond)
	assert.Equal(s.T(), int64(0), zCount(s.T(), s.redisClient, "matchmaking"))
}

type cCounter[T comparable] struct {
	sync.Mutex
	counter map[T]int
}

func NewCCounter[T comparable]() cCounter[T] {
	return cCounter[T]{
		Mutex:   sync.Mutex{},
		counter: make(map[T]int),
	}
}

func (l *cCounter[T]) Increment(item T) {
	l.Lock()
	defer l.Unlock()
	l.counter[item]++
}

func (s *MatchmakingTestSuite) TestMatchmaking_JoinWithManyLobbies() {
	counter := NewCCounter[string]()
	uCounter := NewCCounter[int64]()

	var wg sync.WaitGroup
	testJoin := func(id int64) {
		wg.Add(1)
		go func() {
			lobby, _, err := s.mm.Join(s.ctx, id, s.timeout)
			assert.NoError(s.T(), err)
			counter.Increment(lobby.ID)
			wg.Done()
		}()
	}

	st := time.Now()
	for i := 0; i < maxLobbySize*1000; i++ {
		testJoin(int64(i) + 1)
	}

	wg.Wait()
	fmt.Println("Took", time.Since(st))

	assert.Len(s.T(), counter.counter, 1000)

	// Each user must have joined one lobby
	for lobbyID, count := range counter.counter {
		lobby, err := s.lobby.Get(context.Background(), entity.NewID("lobby", lobbyID))
		assert.NoError(s.T(), err)
		assert.Len(s.T(), lobby.Participants, maxLobbySize)
		assert.Equal(s.T(), count, maxLobbySize)
		for _, participant := range lobby.Participants {
			uCounter.Increment(participant)
		}
	}
	for _, count := range uCounter.counter {
		assert.Equal(s.T(), 1, count)
	}

	// check whether account's current game lobby is ready
	acc, err := s.account.Get(context.Background(), entity.NewID("account", 50))
	assert.NoError(s.T(), err)
	assert.NotEqual(s.T(), "", acc.CurrentLobby)
}

func zCount(t *testing.T, redisClient rueidis.Client, key string) int64 {
	count, err := redisClient.Do(context.Background(), redisClient.B().Zcount().Key(key).Min("-inf").Max("+inf").Build()).ToInt64()
	assert.NoError(t, err)
	return count
}

func redisKeys(t *testing.T, redisClient rueidis.Client, pattern string) []string {
	keys, err := redisClient.Do(context.Background(), redisClient.B().Keys().Pattern(pattern).Build()).AsStrSlice()
	assert.NoError(t, err)
	return keys
}
