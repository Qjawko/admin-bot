package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/qjawko/admin-bot/types"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type ChallengeGenerator interface {
	Gen(ctx context.Context) (types.Challenge, error)
}

type ChallengeRepository interface {
	Get(ctx context.Context, userId int) (types.Challenge, error)
	Set(ctx context.Context, userId int, challenge *types.Challenge) error
	Del(ctx context.Context, userId int) error
}

type ReminderQueue interface {
	Send(ctx context.Context, chatId int64, delay time.Duration) error
}

type Start struct {
	challengeGen        ChallengeGenerator
	challengeRepository ChallengeRepository
	reminderQueue       ReminderQueue
}

func NewStart(challengeGen ChallengeGenerator, challengeStorage ChallengeRepository, queue ReminderQueue) *Start {
	return &Start{challengeGen: challengeGen, challengeRepository: challengeStorage, reminderQueue: queue}
}

func (s Start) ServeCall(ctx context.Context, req *tgbotapi.Message, rsp *tgbotapi.MessageConfig) {
	if req.From == nil && req.From.IsBot {
		rsp.Text = "from field should not be nil and not bot"
		return
	}

	challenge, err := s.challengeGen.Gen(ctx)
	if err != nil {
		rsp.Text = fmt.Sprintf("initializing a challenge: %s", err)
		return
	}

	if err = s.challengeRepository.Set(ctx, req.From.ID, &challenge); err != nil {
		rsp.Text = fmt.Sprintf("saving challenge to repository: %s", err)
		return
	}

	if err = s.reminderQueue.Send(ctx, req.Chat.ID, 10*time.Minute); err != nil {
		rsp.Text = fmt.Sprintf("sending reminder: %s", err)
		return
	}

	rsp.Text = challenge.Question
}
