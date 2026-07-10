package matchmaking

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/0xpelamar/kingscomp/pkg/jsonhelper"
	"github.com/0xpelamar/kingscomp/pkg/randhelper"
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
	lobby             repository.Lobby
	account           repository.Account
	question          repository.Question
}

func NewRedisMatchMaking(client rueidis.Client, lobby repository.Lobby, question repository.Question) *RedisMatchMaking {
	script := rueidis.NewLuaScript(matchMakingScript)
	return &RedisMatchMaking{
		client:            client,
		matchMakingScript: script,
		lobby:             lobby,
		question:          question,
	}
}

func (r RedisMatchMaking) Join(ctx context.Context, userID int64, timeout time.Duration) (entity.Lobby, bool, error) {
	defer func() {
		removeFromQueue := r.client.B().Zrem().Key("matchmaking").Member(strconv.FormatInt(userID, 10)).Build()
		r.client.Do(ctx, removeFromQueue)
	}()

	resp, err := r.matchMakingScript.Exec(ctx, r.client,
		[]string{"matchmaking"},
		[]string{fmt.Sprint(MaxLobbyMembers),
			strconv.FormatInt(time.Now().Add(-time.Minute*2).Unix(), 10),
			uuid.New().String(),
			strconv.FormatInt(userID, 10),
			strconv.FormatInt(time.Now().Unix(), 10),
		}).ToArray()
	if err != nil {
		logrus.WithError(err).Errorln("could not join the matchmaking")
		return entity.Lobby{}, false, err
	}

	// inside a queue
	if len(resp) == 1 {
		//logrus.WithField("userId", userID).Info("waiting for a lobby")
		cmd := r.client.B().Brpop().Key(fmt.Sprintf("matchmaking:%d", userID)).Timeout(timeout.Seconds()).Build()
		result, err := r.client.Do(ctx, cmd).AsStrSlice()
		if err != nil {
			if errors.Is(err, rueidis.Nil) {
				return entity.Lobby{}, false, ErrTimeout
			}
			logrus.WithError(err).Errorln("could not get matchmaking notice from redis")
			return entity.Lobby{}, false, err
		}
		if len(result) < 2 {
			return entity.Lobby{}, false, ErrTimeout
		}
		lobby, err := r.lobby.Get(ctx, entity.NewID("lobby", result[1]))
		return lobby, false, err
	}

	// just created a lobby
	if len(resp) == 3 {
		lobbyID, _ := resp[1].ToString()
		matchedUsers, _ := resp[2].AsIntSlice()

		// create a new lobby
		lobby, err := r.createNewLobby(ctx, lobbyID, matchedUsers)
		if err != nil {
			return entity.Lobby{}, false, err
		}

		return lobby, true, err
	}
	logrus.WithError(err).Errorln("bad redis response")
	return entity.Lobby{}, false, ErrBadRedisResponse
}

func (r RedisMatchMaking) Leave(ctx context.Context, userID int64) error {
	//TODO implement me
	panic("implement me")
}

func (r RedisMatchMaking) createNewLobby(ctx context.Context, lobbyID string, users []int64) (entity.Lobby, error) {
	cmds := make([]rueidis.Completed, 0, 5)
	// get lobby questions
	activeQuestionsCount, err := r.question.GetActiveQuestionsCount(ctx)
	if err != nil {
		return entity.Lobby{}, err
	}
	questionIndexes := randhelper.GenerateDistinctNumbers(LobbyQuestionCount, 0, activeQuestionsCount)

	questions, err := r.question.GetActiveQuestions(ctx, questionIndexes...)
	if err != nil {
		return entity.Lobby{}, err
	}
	// create the lobby
	userStates := make(map[int64]entity.UserState, len(users))
	for _, user := range users {
		userStates[user] = entity.UserState{}
	}
	lobby := entity.Lobby{
		ID:            lobbyID,
		Participants:  users,
		CreatedAtUnix: 0,
		State:         "created",
		UserState:     userStates,
		Questions:     questions,
	}

	cmds = append(cmds,
		r.client.B().JsonSet().
			Key(entity.NewID("lobby", lobby.ID).String()).Path(".").
			Value(string(jsonhelper.Encode(lobby))).Build(),
	)

	// update participants current lobby
	for _, participant := range users {
		userMatchmakingListKey := fmt.Sprintf("matchmaking:%d", participant)
		cmds = append(cmds,
			r.client.B().JsonSet().
				Key(entity.NewID("account", participant).String()).
				Path("$..current_lobby").
				Value(fmt.Sprintf(`"%s"`, lobbyID)).Build(),
			r.client.B().Rpush().Key(userMatchmakingListKey).
				Element(lobbyID).Build(),
			r.client.B().Expire().Key(userMatchmakingListKey).Seconds(120).Build(),
		)
	}
	resp := r.client.DoMulti(ctx, cmds...)
	err = repository.ReduceRedisResponseError(resp, rueidis.Nil)
	if err != nil {
		logrus.WithError(err).Errorln("could not create the matchmaking lobby")
		return entity.Lobby{}, err
	}
	return lobby, nil
}
