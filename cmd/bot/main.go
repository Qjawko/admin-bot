package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qjawko/admin-bot/generator"
	"github.com/qjawko/admin-bot/handlers"
	"github.com/qjawko/admin-bot/internal/queue"
	"github.com/qjawko/admin-bot/repository"
	"github.com/qjawko/admin-bot/router"

	"github.com/go-redis/redis/v8"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/streadway/amqp"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	debug := declareDebugMode()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_BOT_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}
	bot.Debug = debug

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	conn, err := amqp.Dial(os.Getenv("TG_BOT_AMQP_URL"))
	if err != nil {
		log.Fatalln(err)
	}

	tctx, tcancel := context.WithTimeout(ctx, 10*time.Second)
	defer tcancel()
	if err = rdb.Ping(tctx).Err(); err != nil {
		log.Fatalln(fmt.Errorf("ping redis: %w", err))
	}

	challengeRepository := repository.NewChallenge(rdb)
	subscriptionRepository := repository.NewSubscription(rdb)

	routingKey := os.Getenv("TG_BOT_AMQP_ROUTING_KEY")
	queueName := os.Getenv("TG_BOT_AMQP_QUEUE")
	reminderQueue, err := queue.NewReminder(conn, routingKey)
	if err != nil {
		log.Fatalln(err)
	}

	reminderConsumer, err := queue.NewReminderConsumer(logger, &reminderQueue, conn, bot, routingKey, queueName)
	if err != nil {
		log.Fatalln(err)
	}

	startHandler := handlers.NewStart(generator.NewChallenge(), challengeRepository, subscriptionRepository, &reminderQueue)
	nonCommandHandler := handlers.NewDefault(challengeRepository, subscriptionRepository)

	r := router.NewRouter(logger, bot)

	r.SetNonCommandHandler(nonCommandHandler)
	r.Add("/start", startHandler)

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return reminderConsumer.ConsumeMessages(ctx)
	})

	group.Go(func() error {
		return r.ListenAndServe(ctx, tgbotapi.UpdateConfig{})
	})

	slog.Info("bot shutdown")
}

func declareDebugMode() bool {
	var debugBotMode bool
	flag.BoolVar(&debugBotMode, "debugBotMode", true, "")
	flag.Parse()
	return debugBotMode
}
