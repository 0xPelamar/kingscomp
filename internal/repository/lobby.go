package repository

import (
	"context"
	"fmt"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/pkg/jsonhelper"
	"github.com/redis/rueidis"
)

var _ Lobby = &LobbyRedisRepository{}

type LobbyRedisRepository struct {
	*RedisCommonBehaviour[entity.Lobby]
}

func NewLobbyRedisRepository(client rueidis.Client) *LobbyRedisRepository {
	return &LobbyRedisRepository{
		NewRedisCommonBehaviour[entity.Lobby](client),
	}
}

func (l *LobbyRedisRepository) UpdateUserState(ctx context.Context,
	lobbyID string, userID int64, key string, val any) error {
	updatePath := fmt.Sprintf("$.user_state.%d.%s", userID, key)
	cmd := l.client.B().JsonSet().Key(entity.NewID("lobby", lobbyID).String()).
		Path(updatePath).
		Value(string(jsonhelper.Encode(val))).Build()

	return l.client.Do(ctx, cmd).Error()
}
