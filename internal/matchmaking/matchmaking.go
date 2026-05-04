package matchmaking

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/google/uuid"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"

	"time"
)

var (
	ErrBadRedisResponse = errors.New("bad redis response")
)

//go:embed matchmaking.lua
var matchMakingScript string

type MatchMaking interface {
	Join(ctx context.Context, userID int64) (entity.Lobby, error)
	Leave(ctx context.Context, userID int64) error
}

var _ MatchMaking = &RedisMatchMaking{}

type RedisMatchMaking struct {
	client            rueidis.Client
	matchMakingScript *rueidis.Lua
}

func NewRedisMatchMaking(client rueidis.Client) *RedisMatchMaking {
	script := rueidis.NewLuaScript(matchMakingScript)
	return &RedisMatchMaking{
		client:            client,
		matchMakingScript: script,
	}
}
func (r RedisMatchMaking) Join(ctx context.Context, userID int64) (entity.Lobby, error) {
	go r.client.Receive(ctx, r.client.B().Subscribe().Channel("matchmakin").Build(), func(msg rueidis.PubSubMessage) {
		fmt.Println(msg)
	})
	resp, err := r.matchMakingScript.Exec(ctx, r.client,
		[]string{"matchmaking", "matchmaking"},
		[]string{"5",
			strconv.FormatInt(time.Now().Add(-time.Minute*2).Unix(), 10),
			uuid.New().String(),
			strconv.FormatInt(userID, 10),
			strconv.FormatInt(time.Now().Unix(), 10),
		}).ToArray()
	if err != nil {
		logrus.WithError(err).Errorln("could not join the matchmaking")
		return entity.Lobby{}, err
	}

	// inside a queue, we must listen to the pub/sub
	if len(resp) == 1 {

	}
	// just created a lobby
	if len(resp) == 3 {

	}
	logrus.WithError(err).Errorln("bad redis response")
	return entity.Lobby{}, ErrBadRedisResponse
}

func (r RedisMatchMaking) Leave(ctx context.Context, userID int64) error {
	//TODO implement me
	panic("implement me")
}
