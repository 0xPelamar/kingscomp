package lobby

import (
	"github.com/redis/rueidis"
)

type Manager struct {
	client rueidis.Client
}

func NewManager(client rueidis.Client) *Manager {
	return &Manager{
		client: client,
	}
}
