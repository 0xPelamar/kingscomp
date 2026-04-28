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
	"github.com/stretchr/testify/assert"
)

type TestType struct {
	ID   string
	Name string
}

func (t TestType) EntityID() entity.ID {
	return entity.NewID("testType", t.ID)
}

func TestRedisCommonBehaviour_GetAndSave(t *testing.T) {
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

	rcb := repository.NewRedisCommonBehaviour[TestType](redisClient)
	ctx := context.Background()

	err = rcb.Save(ctx, TestType{"21", "asdf"})
	assert.NoError(t, err)

	err = rcb.Save(ctx, TestType{"22", "qwer"})
	assert.NoError(t, err)

	ins, err := rcb.Get(ctx, entity.NewID("testType", "21"))
	assert.NoError(t, err)
	assert.Equal(t, "asdf", ins.Name)
	assert.Equal(t, "21", ins.ID)

	ins, err = rcb.Get(ctx, entity.NewID("testType", "22"))
	assert.NoError(t, err)
	assert.Equal(t, "qwer", ins.Name)
	assert.Equal(t, "22", ins.ID)

	ins, err = rcb.Get(ctx, entity.NewID("testType", "23"))
	assert.ErrorIs(t, err, repository.ErrNotFound)

}
