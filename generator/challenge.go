package generator

import (
	"context"

	"github.com/qjawko/admin-bot/types"
)

type Challenge struct{}

func NewChallenge() *Challenge {
	return &Challenge{}
}

func (c Challenge) Gen(_ context.Context) (types.Challenge, error) {
	return types.Challenge{
		Question: "1+1",
		Answer:   "2",
	}, nil
}
