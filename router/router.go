package router

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type HandlerFunc func(ctx context.Context, req *tgbotapi.Message, rsp *tgbotapi.MessageConfig)

func (f HandlerFunc) ServeCall(ctx context.Context, req *tgbotapi.Message, rsp *tgbotapi.MessageConfig) {
	f(ctx, req, rsp)
}

type Handler interface {
	ServeCall(ctx context.Context, req *tgbotapi.Message, rsp *tgbotapi.MessageConfig)
}

type Router struct {
	log               *slog.Logger
	bot               *tgbotapi.BotAPI
	routes            map[string]Handler
	nonCommandHandler atomic.Pointer[Handler]
}

func (r *Router) SetNonCommandHandler(h Handler) {
	r.nonCommandHandler.Store(&h)
}

func NewRouter(log *slog.Logger, bot *tgbotapi.BotAPI) *Router {
	return &Router{
		bot:               bot,
		log:               log,
		routes:            make(map[string]Handler),
		nonCommandHandler: atomic.Pointer[Handler]{},
	}
}

func (r *Router) Add(command string, h Handler) { // todo: add concurrency safety
	r.routes[command] = h
}

func (r *Router) ListenAndServe(ctx context.Context, config tgbotapi.UpdateConfig) error {
	updates, err := r.bot.GetUpdatesChan(config)
	if err != nil {
		return fmt.Errorf("get updates channel: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			r.bot.StopReceivingUpdates()
			return fmt.Errorf("context canceled: %w", ctx.Err())
		case update, ok := <-updates:
			if !ok {
				return fmt.Errorf("update channel is closed: %w", io.EOF)
			}

			if update.Message == nil {
				r.log.LogAttrs(ctx, slog.LevelDebug, "nil message", slog.Int("update_id", update.UpdateID))
				continue
			}

			if !update.Message.IsCommand() {
				handler := r.nonCommandHandler.Load()
				if handler != nil {
					go r.serveCall(ctx, *handler, update.Message)
				}
				continue
			}

			handler, ok := r.routes[update.Message.Command()]
			if !ok {
				reply := tgbotapi.NewMessage(update.Message.Chat.ID, "command not found")
				_, err = r.bot.Send(reply)
				if err != nil {
					r.log.LogAttrs(ctx, slog.LevelError, "", slog.Int64("chat_id", update.Message.Chat.ID))
					continue
				}
			}

			go r.serveCall(ctx, handler, update.Message)
		}
	}
}

func (r *Router) serveCall(ctx context.Context, handler Handler, msg *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, "")
	handler.ServeCall(ctx, msg, &reply)

	_, err := r.bot.Send(reply)
	if err != nil {
		r.log.LogAttrs(
			ctx,
			slog.LevelError,
			"send error",
			slog.Int64("chat_id", msg.Chat.ID),
		)
		return
	}
}
