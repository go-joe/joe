package slack

import (
	"context"
	"strings"

	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/fgrosse/joe"
)

type Config struct {
	Token  string
	Debug  bool
	Logger *zap.Logger
}

type adapter struct {
	context  context.Context
	logger   *zap.Logger
	client   *slack.Client
	rtm      *slack.RTM
	userID   string
	messages chan joe.Message
}

func Adapter(token string, opts ...Option) joe.Option {
	return func(b *joe.Bot) error {
		conf := Config{Token: token}
		for _, opt := range opts {
			err := opt(&conf)
			if err != nil {
				return err
			}
		}

		if conf.Logger == nil {
			conf.Logger = b.Logger
		}

		a, err := NewAdapter(b.Context, conf)
		if err != nil {
			return err
		}

		b.Adapter = a
		return nil
	}
}

func NewAdapter(ctx context.Context, conf Config) (joe.Adapter, error) {
	a := &adapter{
		client:   slack.New(conf.Token, slack.OptionDebug(conf.Debug)),
		context:  ctx,
		logger:   conf.Logger,
		messages: make(chan joe.Message, 10),
	}

	if a.logger == nil {
		a.logger = zap.NewNop()
	}

	resp, err := a.client.AuthTestContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "slack auth test failed")
	}

	a.userID = resp.UserID
	a.rtm = a.client.NewRTM()

	// Start message handling in two goroutines. They will be closed when we
	// disconnect the RTM upon adapter.Close().
	go a.rtm.ManageConnection()
	go a.handleSlackEvents()

	a.logger.Info("Connected to slack API",
		zap.String("url", resp.URL),
		zap.String("user", resp.User),
		zap.String("user_id", resp.UserID),
		zap.String("team", resp.Team),
		zap.String("team_id", resp.TeamID),
	)

	return a, nil
}

func (a *adapter) handleSlackEvents() {
	for msg := range a.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.MessageEvent:
			a.logger.Debug("Received message", zap.Any("event", ev))
			a.handleMessageEvent(ev)

		case *slack.RTMError:
			a.logger.Error("Slack Real Time Messaging (RTM) error", zap.Any("event", ev))

		case *slack.InvalidAuthEvent:
			a.logger.Error("Invalid authentication error", zap.Any("event", ev))
			return

		default:
			// Ignore other events..
		}
	}

}

func (a *adapter) handleMessageEvent(ev *slack.MessageEvent) {
	// check if we have a DM, or standard channel post
	direct := strings.HasPrefix(ev.Msg.Channel, "D")
	if !direct && !strings.Contains(ev.Msg.Text, "<@"+a.userID+">") {
		// msg not for us!
		return
	}

	text := strings.TrimSpace(strings.TrimPrefix(ev.Text, "<@"+a.userID+">"))
	a.messages <- joe.Message{Text: text, ChannelID: ev.Channel}
}

func (a *adapter) NextMessage() <-chan joe.Message {
	// Replace C2147483705 with your Channel ID
	// rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "C2147483705"))
	return a.messages
}

func (a *adapter) Send(text, channelID string) error {
	a.logger.Info("Sending message to channel",
		zap.String("channel_id", channelID),
		// do not leak actual message content since it might be sensitive
	)

	a.rtm.SendMessage(a.rtm.NewOutgoingMessage(text, channelID))
	return nil
}

func (a *adapter) Close() error {
	return a.rtm.Disconnect()
}
