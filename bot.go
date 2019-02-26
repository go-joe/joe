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
	Logger  *zap.Logger

	initErr error // any error when we created a new bot
}

type Module func(*Config) error

func New(name string, modules ...Module) *Bot {
	ctx := cli.Context()
	logger := NewLogger()
	brain := NewBrain(logger.Named("brain"))

	conf := &Config{
		Context:        ctx,
		Name:           name,
		HandlerTimeout: brain.handlerTimeout,
		logger:         logger,
		adapter:        NewCLIAdapter(ctx, name),
		brain:          brain,
	}

	conf.logger.Info("Initializing bot", zap.String("name", name))
	for _, mod := range modules {
		err := mod(conf)
		if err != nil {
			conf.errs = append(conf.errs, err)
		}
	}

	// apply all configuration options
	brain.handlerTimeout = conf.HandlerTimeout
	return &Bot{
		Name:    conf.Name,
		Context: conf.Context,
		Logger:  conf.logger,
		Adapter: conf.adapter,
		Brain:   brain,
		initErr: multierr.Combine(conf.errs...),
	}
}

func (b *Bot) Run() error {
	if b.initErr != nil {
		return errors.Wrap(b.initErr, "failed to initialize bot")
	}

	if len(b.Brain.registrationErrs) > 0 {
		return multierr.Combine(b.Brain.registrationErrs...)
	}

	b.Adapter.Register(b.Brain)
	b.Brain.Emit(InitEvent{})

	b.Logger.Info("Bot initialized and ready to operate", zap.String("name", b.Name))
	b.Brain.HandleEvents(b.Context)

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

	b.Brain.RegisterHandler(func(ctx context.Context, evt ReceiveMessageEvent) error {
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
