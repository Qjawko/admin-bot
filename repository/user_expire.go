package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Subscription struct {
	rdb redis.UniversalClient
}

func NewSubscription(rdb redis.UniversalClient) *Subscription {
	return &Subscription{rdb: rdb}
}

func (s Subscription) Set(ctx context.Context, expiration time.Duration, userId int) error {
	subscriptionKey := fmt.Sprintf("user:%d:subscription", userId)

	err := s.rdb.Set(ctx, subscriptionKey, "active", expiration).Err()
	return err
}

func (s Subscription) Get(ctx context.Context, userId int) (string, error) {
	subscriptionKey := fmt.Sprintf("user:%d:subscription", userId)

	subscriptionStatus := s.rdb.Get(ctx, subscriptionKey)

	switch {
	case subscriptionStatus.Val() != "":
		return "active", nil
	case errors.Is(subscriptionStatus.Err(), redis.Nil):
		return "deleted", nil
	case subscriptionStatus.Err() != nil:
		return "", subscriptionStatus.Err()
	}

	return subscriptionStatus.Val(), subscriptionStatus.Err()
}
