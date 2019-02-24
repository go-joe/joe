package joe

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/fraugster/cli"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type Bot struct {
	Context context.Context
	Name    string
	Adapter Adapter
	Brain   *Brain
	Events  *EventProcessor
	Logger  *zap.Logger

	initErr error // any error when we created a new bot
}

func New(name string, opts ...Option) *Bot {
	conf := Config{
		Context: cli.Context(),
		Logger:  NewLogger(),
		Name:    name,
	}

	conf.Logger.Info("Initializing bot", zap.String("name", name))
	conf.ApplyOptions(opts)

	events := NewEventProcessor(conf.Logger, conf.HandlerTimeout)

	return &Bot{
		Name:    conf.Name,
		Context: conf.Context,
		Logger:  conf.Logger,
		Adapter: conf.Adapter,
		Brain:   NewBrain(conf.Memory, conf.Logger.Named("brain"), events),
		Events:  events,
		initErr: multierr.Combine(conf.errs...),
	}
}

func (b *Bot) Run() error {
	if b.initErr != nil {
		return errors.Wrap(b.initErr, "failed to initialize bot")
	}

	if len(b.Events.registrationErrs) > 0 {
		return multierr.Combine(b.Events.registrationErrs...)
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
			Context: ctx,
			Text:    evt.Text,
			Channel: evt.Channel,
			Matches: matches[1:],
			adapter: b.Adapter,
		})
	})
}

func (b *Bot) Say(channel, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	err := b.Adapter.Send(msg, channel)
	if err != nil {
		b.Logger.Error("Failed to send message", zap.Error(err))
	}
}
