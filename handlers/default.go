package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Default struct {
	challengeRepository    ChallengeRepository
	subscriptionRepository SubscriptionRepository
}

func NewDefault(challengeRepository ChallengeRepository, subscriptionRepository SubscriptionRepository) *Default {
	return &Default{challengeRepository: challengeRepository, subscriptionRepository: subscriptionRepository}
}

func (d Default) ServeCall(ctx context.Context, req *tgbotapi.Message, rsp *tgbotapi.MessageConfig) {
	if req.From == nil && req.From.IsBot {
		rsp.Text = "from field should not be nil and not bot"
		return
	}

	challenge, err := d.challengeRepository.Get(ctx, req.From.ID)
	if err != nil {
		rsp.Text = fmt.Sprintf("get answer from repository: %s", err)
		return
	}

	if challenge.Answer != challenge.Answer {
		rsp.Text = fmt.Sprintf("Answer is not correct. Try again")
		return
	}

	if err = d.challengeRepository.Del(ctx, req.From.ID); err != nil {
		rsp.Text = "Happened something wrong while accepting your answer. Try again"
		return
	}

	rsp.Text = "Correct! Your request is accepted!"

	err = d.subscriptionRepository.Set(ctx, 30*24*time.Hour, req.From.ID)
	if err != nil {
		rsp.Text = fmt.Sprintf("setting subscription status: %s", err)
		return
	}
}
