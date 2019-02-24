package joe

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fraugster/cli"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Bot struct {
	Context context.Context
	Name    string
	Adapter Adapter
	Brain   Brain
	Events  *EventProcessor
	Logger  *zap.Logger

	initErr error // any error when we created a new bot
}

func New(name string, opts ...Option) *Bot {
	b := &Bot{
		Context: cli.Context(),
		Logger:  NewLogger(),
		Name:    name,
	}

	handlerTimeout := 10 * time.Second // TODO: should be configurable via options just as the logger
	b.Events = NewEventProcessor(b.Logger, handlerTimeout)

	b.Logger.Info("Initializing bot", zap.String("name", name))

	for _, opt := range opts {
		err := opt(b)
		if err != nil && b.initErr == nil {
			b.initErr = err
		}
	}

	if b.Adapter == nil {
		b.Adapter = NewCLIAdapter(b.Context, name)
	}

	if b.Brain == nil {
		b.Brain = NewInMemoryBrain()
	}

	return b
}

func (b *Bot) Run() error {
	if b.initErr != nil {
		return errors.Wrap(b.initErr, "failed to initialize bot")
	}

	b.Adapter.Register(b.Events)
	b.Events.Emit(InitEvent{})

	b.Logger.Info("Bot initialized and ready to operate", zap.String("name", b.Name))
	b.Events.Process(b.Context)

	err := b.Adapter.Close()
	b.Logger.Info("Bot is shutting down", zap.String("name", b.Name))
	if err != nil {
		b.Logger.Info("Error while closing adapter", zap.Error(err))
	}

	return nil
}

type RespondFunc func(Message) error

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

	regex, err := regexp.Compile(expr)
	if err != nil {
		b.Logger.Error("Failed to add Response handler", zap.Error(err))
		return
	}

	b.Events.RegisterHandler(func(ctx context.Context, evt ReceiveMessageEvent) error {
		matches := regex.FindStringSubmatch(evt.Text)
		if len(matches) == 0 {
			return nil
		}

		return fun(Message{
			Context:   ctx,
			Text:      evt.Text,
			ChannelID: evt.ChannelID,
			Matches:   matches[1:],
			adapter:   b.Adapter,
		})
	})
}

func (b *Bot) Say(channelID, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	err := b.Adapter.Send(msg, channelID)
	if err != nil {
		b.Logger.Error("Failed to send message", zap.Error(err))
	}
}
