package redis

import (
	"context"
	"fmt"

	"user-management-service/internal/config"

	"github.com/redis/go-redis/v9"
)

type Cash struct {
	client *redis.Client
}

func New(cfg config.Cache) (*Cash, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		DB:   cfg.DB,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	err = client.SAdd(ctx, "blacklist", "").Err()
	if err != nil {
		return nil, err
	}

	return &Cash{client: client}, nil
}

func (c *Cash) AddToBlaclist(ctx context.Context, token string) error {
	const op = "SearchInBlacklist"

	err := c.client.SAdd(ctx, "blacklist", token).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Cash) SearchInBlacklist(ctx context.Context, token string) (bool, error) {
	const op = "SearchInBlacklist"

	found, err := c.client.SIsMember(ctx, "blacklist", token).Result()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return found, nil
}
