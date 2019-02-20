package joe

import (
	"context"
	"fmt"
	"strings"

	"github.com/fraugster/cli"
	"go.uber.org/zap"
)

type Bot struct {
	Context context.Context
	Adapter Adapter
	Brain   Brain
	Logger  *zap.Logger
	Name    string

	handlers []responseHandler
}

// TODO: can use options patters to select a logger or adapter
func New(name string) *Bot {
	ctx := cli.Context()
	logger := NewLogger()

	return &Bot{
		Context: ctx,
		Adapter: NewCLIAdapter(ctx, name),
		Brain:   NewInMemoryBrain(),
		Logger:  logger,
		Name:    name,
	}
}

func (b *Bot) Run() {
	b.Logger.Info("Started bot", zap.String("name", b.Name))
	for {
		select {
		case msg := <-b.Adapter.NextMessage():
			b.handleMessage(msg)

		case <-b.Context.Done():
			err := b.Adapter.Close()
			b.Logger.Info("Bot is shutting down", zap.String("name", b.Name))
			if err != nil {
				b.Logger.Info("Error while closing adapter", zap.Error(err))
			}
			return
		}
	}
}

func (b *Bot) handleMessage(s string) {
	msg := Message{
		Context: b.Context,
		Msg:     s,
	}

	for _, h := range b.handlers {
		matches := h.regex.FindStringSubmatch(s)
		if len(matches) == 0 {
			continue
		}

		msg.Matches = matches[1:]
		err := h.run(msg)
		if err != nil {
			b.Logger.Error("Failed to handle message", zap.Error(err))
		} else {
			// return after first match
			return
		}
	}
}

func (b *Bot) Respond(msg string, fun RespondFunc) {
	expr := "^" + msg + "$"
	b.RespondRegex(expr, fun)
}

func (b *Bot) RespondRegex(expr string, fun RespondFunc) {
	if expr == "" {
		return
	}

	if expr[0] == '^' {
		// String starts with the "^" anchor but does it also have the prefix
		// or case insensitive matching?
		if !strings.HasPrefix(expr, "^(?i)") {
			expr = "^(?i)" + expr[1:]
		}
	} else {
		// The string is not starting with "^" but maybe it has the prefix for
		// case insensitive matching already?
		if !strings.HasPrefix(expr, "(?i)") {
			expr = "(?i)" + expr
		}
	}

	h, err := newHandler(expr, fun)
	if err != nil {
		b.Logger.Fatal("Failed to add Response handler", zap.Error(err))
	}

	b.handlers = append(b.handlers, h)
}

func (b *Bot) Say(msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	err := b.Adapter.Send(msg)
	if err != nil {
		b.Logger.Error("Failed to send message", zap.Error(err))
	}
}
