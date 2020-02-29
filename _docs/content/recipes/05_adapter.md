+++
title = "Implement a new Adapter"
slug = "adapter"
weight = 5
+++

Adapters let you interact with the outside world by receiving and sending messages.
Joe currently has the following seven Adapter implementations:

- <i class="fas fa-terminal"></i> CLI Adapter: https://github.com/go-joe/joe
- <i class='fab fa-slack fa-fw'></i> Slack Adapter: https://github.com/go-joe/slack-adapter
- <i class='fab fa-rocketchat fa-fw'></i> Rocket.Chat Adapter: https://github.com/dwmunster/rocket-adapter
- <i class='fab fa-telegram fa-fw'></i> Telegram Adapter: https://github.com/robertgzr/joe-telegram-adapter
- <i class='fas fa-hashtag fa-fw'></i> IRC Adapter: https://github.com/akrennmair/joe-irc-adapter
- <i class="fas fa-circle-notch"></i> Mattermost Adapter: https://github.com/dwmunster/joe-mattermost-adapter
- <i class='fab fa-vk'></i> VK Adapter: https://github.com/tdakkota/joe-vk-adapter 

If you want to integrate with a chat service that is not listed above, you can
write your own Adapter implementation.

### Adapters are Modules

Firstly, your adapter should be available as [`joe.Module`][module] so it can
easily be integrated into the bot via the [`joe.New(…)`][new] function.

The `Module` interface looks like this:

```go
// A Module is an optional Bot extension that can add new capabilities such as
// a different Memory implementation or Adapter.
type Module interface {
	Apply(*Config) error
}
```

To easily implement a Module without having to declare an `Apply` function on
your chat adapter type, you can use the `joe.ModuleFunc` type. For instance the
Slack adapter uses the following, to implement it's `Adapter(…)` function:

```go
// Adapter returns a new Slack adapter as joe.Module.
//
// Apart from the typical joe.ReceiveMessageEvent event, this adapter also emits
// the joe.UserTypingEvent. The ReceiveMessageEvent.Data field is always a
// pointer to the corresponding github.com/nlopes/slack.MessageEvent instance.
func Adapter(token string, opts ...Option) joe.Module {
	return joe.ModuleFunc(func(joeConf *joe.Config) error {
		conf, err := newConf(token, joeConf, opts)
		if err != nil {
			return err
		}

		a, err := NewAdapter(joeConf.Context, conf)
		if err != nil {
			return err
		}

		joeConf.SetAdapter(a)
		return nil
	})
}
```

The passed `*joe.Config` parameter can be used to lookup general options such as
the `context.Context` used by the bot. Additionally you can create a named
logger via the `Config.Logger(…)` function and you can register extra handlers
or [emit events](/recipes/events) via the `Config.EventEmitter()` function.
 
Most importantly for an Adapter implementation however is, that it finally needs
to register itself via the `Config.SetAdapter(…)` function.

By defining an `Adapter(…)` function in your package, it is now possible to use
your adapter as Module passed to `joe.New(…)`. Additionally your `NewAdapter(…)`
function is useful to directly create a new adapter instance which can be used
during unit tests. Last but not least, the options pattern has proven useful in
this kind of setup and is considered good practice when writing modules in general.

### The Adapter Interface

```go
// An Adapter connects the bot with the chat by enabling it to receive and send
// messages. Additionally advanced adapters can emit more events than just the
// ReceiveMessageEvent (e.g. the slack adapter also emits the UserTypingEvent).
// All adapter events must be setup in the RegisterAt function of the Adapter.
//
// Joe provides a default CLIAdapter implementation which connects the bot with
// the local shell to receive messages from stdin and print messages to stdout.
type Adapter interface {
	RegisterAt(*Brain)
	Send(text, channel string) error
	Close() error
}
``` 

The most straight forwards function to implement should be the `Send(…)` and
`Close(…)` functions. The `Send` function should output the given text to the
specified channel as the Bot. The initial connection and authentication to send
these messages should have been setup earlier by your `Adapter` function as
shown above. When the bot shuts down, it will call the `Close()` function of
your adapter so you can terminate your connection and release all resources you
have opened.

In order to also _receive_ messages and pass them to Joe's event handler you
need to implement a `RegisterAt(*joe.Brain)` function. This function gets called
during the setup of the bot and allows the adapter to directly access to the Brain.
This function must not block and thus will typically spawn a new goroutine which
should be stopped when the `Close()` function of your adapter implementation is
called.

In this goroutine you should listen for new messages from your chat application
(e.g. via a callback or polling it). When a new message is received, you need to
emit it as `joe.ReceiveMessageEvent` to the brain.

E.g. for the Slack adapter, this looks like this:

```go
func (a *BotAdapter) handleMessageEvent(ev *slack.MessageEvent, brain *joe.Brain) {
	// Check if the message comes from ourselves.
	if ev.User == a.userID {
		// Message is from us, ignore it!
		return
	}

	// Check if we have a direct message, or standard channel post.
	selfLink := a.userLink(a.userID)
	direct := strings.HasPrefix(ev.Msg.Channel, "D")
	if !direct && !strings.Contains(ev.Msg.Text, selfLink) {
		// Message is not meant for us!
		return
	}

	text := strings.TrimSpace(strings.TrimPrefix(ev.Text, selfLink))
	brain.Emit(joe.ReceiveMessageEvent{
		Text:     text,
		Channel:  ev.Channel,
		ID:       ev.Timestamp, // slack uses the message timestamps as identifiers within the channel
		AuthorID: ev.User,
		Data:     ev,
	})
}
```

In the snippet above you can see some of the common pitfalls:

- the adapter should ignore it's own messages or it risks ending up in an infinitive loop
- the adapter must make sure the message is actually intended for the bot
- maybe the message needs to be trimmed
- you should try and fill all fields of the `joe.ReceiveMessageEvent`

### Optional Interfaces

Currently there is only a single optional interface that can be implemented by an
Adapter, which is the `joe.ReactionAwareAdapter`:

```go
// ReactionAwareAdapter is an optional interface that Adapters can implement if
// they support reacting to messages with emojis.
type ReactionAwareAdapter interface {
	React(reactions.Reaction, Message) error
}
```

This interface is meant for chat adapters that have emoji support to attach
reactions to previously received messages (e.g. :thumbsup: or :robot:).

### Getting Help

Generally writing an adapter should not be very hard but it's a good idea to
look at the other adapter implementations to get a better understanding of how
to implement your own. If you have questions or need help, simply open an
issue at the [Joe repository at GitHub](https://github.com/go-joe/joe/issues/new).  

Happy adaptering :robot::tada:

[module]: https://godoc.org/github.com/go-joe/joe#Module
[new]: https://godoc.org/github.com/go-joe/joe#New
