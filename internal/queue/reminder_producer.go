package queue

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

type ReminderProducer struct {
	ch         *amqp.Channel
	routingKey string
}

func NewReminder(conn *amqp.Connection, routingKey string) (ReminderProducer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return ReminderProducer{}, fmt.Errorf("init amqp channel: %w", err)
	}

	args := amqp.Table{
		"x-delayed-type": "direct",
	}
	err = ch.ExchangeDeclare("reminder-delayed-exchange", "x-delayed-message", true, false,
		false, false, args)
	if err != nil {
		return ReminderProducer{}, fmt.Errorf("declare delayed exchange: %w", err)
	}

	return ReminderProducer{ch: ch, routingKey: routingKey}, nil
}

func (r *ReminderProducer) Send(_ context.Context, chatId int64, delay time.Duration) error {
	body := strconv.FormatInt(chatId, 10)
	headers := amqp.Table{
		"x-delay": delay,
	}

	pub := amqp.Publishing{
		Headers:     headers,
		ContentType: "text/plain",
		Body:        []byte(body),
	}

	err := r.ch.Publish("reminder-delayed-exchange", r.routingKey, false, false, pub)
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}

func (r *ReminderProducer) Close() error {
	return r.ch.Close()
}
