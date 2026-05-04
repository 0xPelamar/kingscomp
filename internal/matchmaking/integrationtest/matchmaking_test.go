package integrationtest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/0xpelamar/kingscomp/internal/matchmaking"
	"github.com/0xpelamar/kingscomp/internal/repository/redis"
	"github.com/ory/dockertest/v4"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
)

func TestMatchMaking(t *testing.T) {
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

	mm := matchmaking.NewRedisMatchMaking(redisClient)
	mm.Join(ctx, 11)
	assert.Equal(t, int64(1), zCount(t, redisClient, "matchmaking"))
	mm.Join(ctx, 12)
	mm.Join(ctx, 13)
	mm.Join(ctx, 14)
	mm.Join(ctx, 15)
	assert.Equal(t, int64(0), zCount(t, redisClient, "matchmaking"))
	// check if the lobby has been created
	keys := redisKeys(t, redisClient, "*")
	assert.Len(t, keys, 1)
	lobbyKey := keys[0]
	assert.Contains(t, lobbyKey, "lobby:")

	fmt.Println(redisClient.Do(ctx, redisClient.B().JsonGet().Key(lobbyKey).Path(".").Build()).ToString())
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
