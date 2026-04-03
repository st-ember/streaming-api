package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	Rdb redis.UniversalClient
}

func NewClient(addrs []string, password string) (*Client, error) {
	rdb := redis.NewUniversalClient(
		&redis.UniversalOptions{
			Addrs:    addrs,
			Password: password,
		},
	)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Client{Rdb: rdb}, nil
}
