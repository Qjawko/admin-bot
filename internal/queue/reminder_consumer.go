package queue

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/streadway/amqp"
)

type ReminderConsumer struct {
	producer   *ReminderProducer
	log        *slog.Logger
	bot        *tgbotapi.BotAPI
	ch         *amqp.Channel
	routingKey string
	queueName  string
}

func NewReminderConsumer(log *slog.Logger, producer *ReminderProducer, conn *amqp.Connection, bot *tgbotapi.BotAPI, routingKey, queueName string) (ReminderConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return ReminderConsumer{}, nil
	}

	args := amqp.Table{
		"x-delayed-type": "direct",
	}
	err = ch.ExchangeDeclare("reminder-delayed-exchange", "x-delayed-message", true, false,
		false, false, args)
	if err != nil {
		return ReminderConsumer{}, fmt.Errorf("declare delayed exchange: %w", err)
	}

	_, err = ch.QueueDeclare(
		"reminder-delayed-queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return ReminderConsumer{}, fmt.Errorf("declare delayed queue: %w", err)
	}

	err = ch.QueueBind("reminder-delayed-queue", routingKey, "reminder-delayed-exchange",
		false, nil)
	if err != nil {
		return ReminderConsumer{}, fmt.Errorf("bind delayed queue: %w", err)
	}

	return ReminderConsumer{log: log, producer: producer, bot: bot, routingKey: routingKey, queueName: queueName}, nil
}

func (r ReminderConsumer) ConsumeMessages(ctx context.Context) error {
	msgs, err := r.ch.Consume("reminder-delayed-queue", "", true, false, false,
		false, nil)
	if err != nil {
		return fmt.Errorf("consume messages: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-msgs:
			chatId, err := strconv.ParseInt(string(msg.Body), 10, 64)
			if err != nil {
				r.log.LogAttrs(ctx, slog.LevelError, "parsing body of delayed message",
					slog.String("body", string(msg.Body)),
					slog.String("error", err.Error()),
				)
				_ = msg.Nack(false, false)
				continue
			}

			delay, ok := msg.Headers["x-delay"].(time.Duration)
			if !ok {
				delay = 10 * time.Minute
			}

			delay *= 2

			tgMsgCfg := tgbotapi.NewMessage(chatId, "Remind you to answer")
			if _, err := r.bot.Send(tgMsgCfg); err != nil {
				r.log.LogAttrs(ctx, slog.LevelError, "send to chat",
					slog.Int64("chat_id", chatId),
					slog.String("error", err.Error()))
				_ = msg.Nack(false, true)
				continue
			}

			if err = r.producer.Send(ctx, chatId, delay); err != nil {
				r.log.LogAttrs(ctx, slog.LevelError, "send new reminder",
					slog.Duration("delay", delay),
					slog.Int64("chat_id", chatId),
					slog.String("error", err.Error()))
				continue
			}
		}
	}
}
