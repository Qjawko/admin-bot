package repository

import (
	"context"
	"fmt"

	"github.com/qjawko/admin-bot/types"

	"github.com/go-redis/redis/v8"
)

// todo: use transactions

type Challenge struct {
	rdb redis.UniversalClient
}

func NewChallenge(rdb redis.UniversalClient) *Challenge {
	return &Challenge{rdb: rdb}
}

func (c Challenge) Set(ctx context.Context, userId int, challenge *types.Challenge) error {
	err := c.rdb.Set(ctx, fmt.Sprintf("%d:challenge", userId), challenge.Question, 0).Err()
	if err != nil {
		return fmt.Errorf("set challenge: %w", err)
	}

	err = c.rdb.Set(ctx, fmt.Sprintf("%d:answer", userId), challenge.Answer, 0).Err()
	if err != nil {
		return fmt.Errorf("set answer: %w", err)
	}

	return nil
}

func (c Challenge) Get(ctx context.Context, userId int) (types.Challenge, error) {
	var challenge types.Challenge

	err := c.rdb.Get(ctx, fmt.Sprintf("%d:challenge", userId)).Scan(&challenge.Question)
	if err != nil {
		return challenge, fmt.Errorf("get challenge: %w", err)
	}

	err = c.rdb.Get(ctx, fmt.Sprintf("%d:answer", userId)).Scan(&challenge.Answer)
	if err != nil {
		return challenge, fmt.Errorf("get answer: %w", err)
	}

	return challenge, nil
}

func (c Challenge) Del(ctx context.Context, userId int) error {
	err := c.rdb.Del(ctx, fmt.Sprintf("%d:challenge", userId)).Err()
	if err != nil {
		return fmt.Errorf("delete challenge: %w", err)
	}

	err = c.rdb.Del(ctx, fmt.Sprintf("%d:answer", userId)).Err()
	if err != nil {
		return fmt.Errorf("delete answer: %w", err)
	}

	return nil
}
