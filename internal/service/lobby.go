package service

import "github.com/0xpelamar/kingscomp/internal/repository"

type LobbyService struct {
	repository.Lobby
}

func NewLobbyService(rep repository.Lobby) *LobbyService {
	return &LobbyService{
		Lobby: rep,
	}
}
