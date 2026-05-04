package matchmaking

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"

	"time"
)

var (
	ErrBadRedisResponse = errors.New("bad redis response")
	ErrTimeout          = errors.New("lobby queue timeout")
)

//go:embed matchmaking.lua
var matchMakingScript string

type MatchMaking interface {
	Join(ctx context.Context, userID int64, timeout time.Duration) (entity.Lobby, bool, error)
	Leave(ctx context.Context, userID int64) error
}

var _ MatchMaking = &RedisMatchMaking{}

type RedisMatchMaking struct {
	client            rueidis.Client
	matchMakingScript *rueidis.Lua
	lobby             repository.LobbyRepository
}

func NewRedisMatchMaking(client rueidis.Client, lobby repository.LobbyRepository) *RedisMatchMaking {
	script := rueidis.NewLuaScript(matchMakingScript)
	return &RedisMatchMaking{
		client:            client,
		matchMakingScript: script,
		lobby:             lobby,
	}
}

type joinLobbyPubSubResponse struct {
	err     error
	lobbyID string
}

func (r RedisMatchMaking) Join(ctx context.Context, userID int64, timeout time.Duration) (entity.Lobby, bool, error) {
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
		return entity.Lobby{}, false, err
	}

	// inside a queue, we must listen to the pub/sub
	if len(resp) == 1 {
		return entity.Lobby{}, false, nil
	}
	// just created a lobby
	if len(resp) == 3 {
		lobbyID, _ := resp[1].ToString()
		lobby, err := r.lobby.Get(ctx, entity.NewID("lobby", lobbyID))

		cmds := make([]rueidis.Completed, 0, 5)
		for _, participant := range lobby.Participants {
			cmds = append(cmds, r.client.B().JsonSet().Key(entity.NewID("account", participant).String()).Path("$.current_lobby").Value(fmt.Sprintf(`"%s"`, lobbyID)).Build())
		}
		_ = r.client.DoMulti(ctx, cmds...)
		return lobby, true, err
	}
	logrus.WithError(err).Errorln("bad redis response")
	return entity.Lobby{}, false, ErrBadRedisResponse
}

func (r RedisMatchMaking) Leave(ctx context.Context, userID int64) error {
	//TODO implement me
	panic("implement me")
}
