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
)

func TestMatchMaking_Join(t *testing.T) {
	pool := dockertest.NewPoolT(t, "")
	redisResource := pool.RunT(t, "redis", dockertest.WithTag("8.4-alpine"))

	redisPort := redisResource.GetPort("6379/tcp")

	err := pool.Retry(t.Context(), 30*time.Second, func() error {
		_, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
		return err
	})
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
	assert.NoError(t, err)
	defer redisClient.Close()

	ctx := context.Background()
	timeout := 10 * time.Second
	mm := matchmaking.NewRedisMatchMaking(redisClient, repository.NewLobbyRedisRepository(redisClient))

	var wg sync.WaitGroup
	testJoin := func(id int64) {
		wg.Add(1)
		go func() {
			lobby, created, err := mm.Join(ctx, id, timeout)
			assert.NoError(t, err)
			if created {
				assert.NotEqual(t, "", lobby.ID)
			}
			wg.Done()
		}()
	}
	testJoin(11)
	testJoin(12)
	testJoin(13)
	testJoin(14)

	<-time.After(500 * time.Millisecond)

	assert.Equal(t, int64(4), zCount(t, redisClient, "matchmaking"))

	lobby, _, err := mm.Join(ctx, 15, timeout)
	assert.NoError(t, err)
	assert.NotEqual(t, "", lobby.ID)
	wg.Wait()

	//// check if the lobby has been created
	//keys := redisKeys(t, redisClient, "*")
	//assert.Len(t, keys, 1)
	//lobbyKey := keys[0]
	//assert.Contains(t, lobbyKey, "lobby:")
	//
	//fmt.Println(redisClient.Do(ctx, redisClient.B().JsonGet().Key(lobbyKey).Path(".").Build()).ToString())
}

type cCounter[T comparable] struct {
	sync.Mutex
	counter map[T]int
}

func newCCounter[T comparable]() cCounter[T] {
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
func TestMatchMaking_JoinWithManyLobbies(t *testing.T) {
	pool := dockertest.NewPoolT(t, "")
	redisResource := pool.RunT(t, "redis", dockertest.WithTag("8.4-alpine"))

	redisPort := redisResource.GetPort("6379/tcp")

	err := pool.Retry(t.Context(), 30*time.Second, func() error {
		_, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
		return err
	})
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("127.0.0.1:%s", redisPort))
	assert.NoError(t, err)
	defer redisClient.Close()

	accountRepository := repository.NewAccountRedisRepository(redisClient)
	accountRepository.Save(context.Background(), entity.Account{
		ID:        100,
		FirstName: "whatever",
	})

	ctx := context.Background()
	timeout := 10 * time.Second
	lobbyRepository := repository.NewLobbyRedisRepository(redisClient)
	mm := matchmaking.NewRedisMatchMaking(redisClient, lobbyRepository)

	lCounter := newCCounter[string]()
	uCounter := newCCounter[int64]()

	var wg sync.WaitGroup
	testJoin := func(id int64) {
		wg.Add(1)
		go func() {
			lobby, created, err := mm.Join(ctx, id, timeout)
			assert.NoError(t, err)
			if created {
				lCounter.Increment(lobby.ID)
			}
			wg.Done()
		}()
	}
	var members = 1000
	s := time.Now()
	for i := 0; i < members; i++ {
		testJoin(int64(i) + 1)
	}
	wg.Wait()
	fmt.Println(time.Since(s))
	assert.Len(t, lCounter.counter, members/5)

	// Each user must have joined one lobby
	for lobbyID, _ := range lCounter.counter {
		lobby, err := lobbyRepository.Get(context.Background(), entity.NewID("lobby", lobbyID))
		assert.NoError(t, err)

		for _, participant := range lobby.Participants {
			uCounter.Increment(participant)
		}
	}
	for _, count := range uCounter.counter {
		assert.Equal(t, 1, count)
	}

	// check whether account's current game lobby is ready
	acc, err := accountRepository.Get(context.Background(), entity.NewID("account", 100))
	assert.NoError(t, err)
	assert.NotEqual(t, "", acc.CurrentLobby)
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
