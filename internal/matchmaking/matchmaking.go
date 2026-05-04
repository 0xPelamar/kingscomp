package matchmaking

import (
	"context"
	_ "embed"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/rueidis"
	"github.com/samber/lo"
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
	Join(ctx context.Context, userID int64, timeout time.Duration) (entity.Lobby, error)
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

func (r RedisMatchMaking) Join(ctx context.Context, userID int64, timeout time.Duration) (entity.Lobby, error) {
	waitingLobbyCtx, lobbyContextCancel := context.WithTimeout(context.Background(), timeout)
	defer lobbyContextCancel()

	responeChannel := make(chan joinLobbyPubSubResponse, 1)
	go r.client.Receive(waitingLobbyCtx, r.client.B().Subscribe().Channel("matchmaking").Build(), func(msg rueidis.PubSubMessage) {
		message := strings.Split(msg.Message, ":")
		lobbyID := message[0]
		users := lo.Map(strings.Split(message[1], ","), func(item string, _ int) int64 {
			id, _ := strconv.ParseInt(item, 10, 64)
			return id
		})
		if !slices.Contains(users, userID) {
			return
		}
		responeChannel <- joinLobbyPubSubResponse{lobbyID: lobbyID}
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
		select {
		case pubSubResponse := <-responeChannel:
			return r.lobby.Get(ctx, entity.NewID("lobby", pubSubResponse.lobbyID))
		case <-waitingLobbyCtx.Done():
			return entity.Lobby{}, ErrTimeout
		}

	}
	// just created a lobby
	if len(resp) == 3 {
		lobbyID, _ := resp[1].ToString()
		return r.lobby.Get(ctx, entity.NewID("lobby", lobbyID))
	}
	logrus.WithError(err).Errorln("bad redis response")
	return entity.Lobby{}, ErrBadRedisResponse
}

func (r RedisMatchMaking) Leave(ctx context.Context, userID int64) error {
	//TODO implement me
	panic("implement me")
}
