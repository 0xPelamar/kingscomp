package service

import "github.com/0xpelamar/kingscomp/internal/repository"

type LobbyService struct {
	Lobby repository.LobbyRepository
}

func NewLobbyService(lobby repository.LobbyRepository) *LobbyService {
	return &LobbyService{
		Lobby: lobby,
	}
}
