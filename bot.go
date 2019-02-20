package joe

import (
	"context"
	"fmt"

	"github.com/fraugster/cli"
	"go.uber.org/zap"
)

type Bot struct {
	Context context.Context
	Adapter Adapter
	Logger  *zap.Logger
	Name    string

	handlers []responseHandler
}

// TODO: can use options patters to select a logger or adapter
func New(name string) *Bot {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	ctx := cli.Context()
	return &Bot{
		Context: ctx,
		Adapter: NewCLIAdapter(ctx, name),
		Logger:  logger,
		Name:    name,
	}
}

func (b *Bot) Run() {
	b.Logger.Info("Started bot", zap.String("name", b.Name))
	for {
		select {
		case <-b.Context.Done():
			return
		case msg := <-b.Adapter.NextMessage():
			b.handleMessage(msg)
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
		}
	}
}

func (b *Bot) Respond(msg string, fun RespondFunc) {
	expr := "^(?i)" + msg + "$"
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
