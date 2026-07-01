package redis

import (
	"os"

	"github.com/redis/rueidis"
)

func NewRedisClient(address string) (rueidis.Client, error) {
	return rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{address},
		Password:    os.Getenv("REDIS_PASSWORD"),
	})
}
