package service

import (
	"context"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/samber/lo"
)

type App struct {
	Account *AccountService
	Lobby   *LobbyService
}

func NewApp(account *AccountService, lobby *LobbyService) *App {
	return &App{
		Account: account,
		Lobby:   lobby,
	}
}

func (a *App) LobbyPlayers(ctx context.Context, lobbyID string) (entity.Lobby, []entity.Account, error) {
	lobby, err := a.Lobby.Lobby.Get(ctx, entity.NewID("lobby", lobbyID))
	if err != nil {
		return entity.Lobby{}, nil, err
	}

	accounts, err := a.Account.Account.Mget(ctx, lo.Map(lobby.Participants, func(item int64, _ int) entity.ID {
		return entity.NewID("account", item)
	})...)
	if err != nil {
		return entity.Lobby{}, nil, err
	}
	return lobby, accounts, nil
}
