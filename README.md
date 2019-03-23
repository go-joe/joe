<h1 align="center">Joe Bot</h1>
<p align="center">A general-purpose bot library inspired by Hubot but written in Go.</p>
<p align="center">
	<a href="https://github.com/go-joe/joe/releases"><img src="https://img.shields.io/github/tag/go-joe/joe.svg?label=version&color=brightgreen"></a>
	<a href="https://circleci.com/gh/go-joe/joe/tree/master"><img src="https://circleci.com/gh/go-joe/joe/tree/master.svg?style=shield"></a>
	<a href="https://goreportcard.com/report/github.com/go-joe/joe"><img src="https://goreportcard.com/badge/github.com/go-joe/joe"></a>
	<a href="https://codecov.io/gh/go-joe/joe"><img src="https://codecov.io/gh/go-joe/joe/branch/master/graph/badge.svg"/></a>
	<a href="https://godoc.org/github.com/go-joe/joe"><img src="https://img.shields.io/badge/godoc-reference-blue.svg?color=blue"></a>
	<a href="https://github.com/go-joe/joe/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-BSD--3--Clause-blue.svg"></a>
</p>

---

Joe is a library used to write chat bots in [the Go programming language][go].
It is very much inspired by the awesome [Hubot][hubot] framework developed by the
folks at Github and brings its power to people who want to implement chat bots using Go.

**THIS SOFTWARE IS STILL IN ALPHA AND THERE ARE NO GUARANTEES REGARDING API STABILITY YET.**

## Getting Started

Joe is packaged using the new [Go modules][go-modules]. Therefore the recommended
installation method is to add joe to your `go.mod` via: 

```
require github.com/go-joe/joe v0.5.0
```

If you do not use modules yet or you want to hack on Joe you can also go get Joe directly:

```bash
go get github.com/go-joe/joe
```

### Minimal example

The simplest chat bot listens for messages on a chat _Adapter_ and then executes
a _Handler_ function if it sees a message directed to the bot that matches a given pattern.

For example a bot that responds to a message "ping" with the answer "PONG" looks like this:

```go
package main

import "github.com/go-joe/joe"

func main() {
	b := joe.New("example-bot")
	b.Respond("ping", Pong)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func Pong(msg joe.Message) error {
	msg.Respond("PONG")
	return nil
}
```

### Useful example 

Each bot consists of a chat _Adapter_ (e.g. to integrate with Slack), a _Memory_
implementation to remember key-value data (e.g. using Redis) and a _Brain_ which
routes new messages or custom events (e.g. receiving an HTTP call) to the
corresponding registered _handler_ functions.

By default `joe.New(…)` uses the CLI adapter which makes the bot read messages
from stdin and respond on stdout. Additionally the brain will store key value
data in-memory which means it will forget anything you told it when it is restarted.
This default setup is useful for local development without any dependencies but
you will quickly want to add other _Modules_ to extend the bots capabilities.

For instance we can extend the previous example to connect the Bot with a Slack
workspace and store key value data in Redis. To allow the message handlers to
access the memory we define them as functions on a custom `ExampleBot`type which
embeds the `joe.Bot`.

```go
package main

import (
	"strings"
	"github.com/go-joe/joe"
	"github.com/go-joe/redis-memory"
	"github.com/go-joe/slack-adapter"
	"github.com/pkg/errors"
)

type ExampleBot struct {
	*joe.Bot
}

func main() {
	b := &ExampleBot{
		Bot: joe.New("example",
			redis.Memory("localhost:6379"),
			slack.Adapter("xoxb-1452345…"),
		),
	}

	b.Respond("remember (.+) is (.+)", b.Remember)
	b.Respond("what is (.+)", b.WhatIs)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func (b *ExampleBot) Remember(msg joe.Message) error {
	key, value := msg.Matches[0], msg.Matches[1]
	msg.Respond("OK, I'll remember %s is %s", key, value)
	return b.Brain.Set(key, value)
}

func (b *ExampleBot) WhatIs(msg joe.Message) error {
	key := msg.Matches[0]
	value, ok, err := b.Brain.Get(key)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve key %q from brain", key)
	}

	if ok {
		msg.Respond("%s is %s", key, value)
	} else {
		msg.Respond("I do not remember %q", key)
	}

	return nil
}
```

### Handling custom events

The previous example should give you an idea already on how to write simple chat
bots. It is missing one important part however: how can a bot trigger any
interaction proactively, i.e. without a message from a user.

To solve this problem, joe's Brain implements an event handler that you can hook
into. In fact the `Bot.Respond(…)` function that we used in the earlier examples
is doing exactly that to listen for any `joe.ReceiveMessageEvent` that match the
specified regular expression and then execute the handler function.

Implementing custom events is easy because you can emit any type as event and
register handlers that match only this type. What this exactly means is best
demonstrated with another example:

```go
package main

import (
	"time"
	"github.com/go-joe/joe"
)

type ExampleBot struct {
	*joe.Bot
	Channel string // example for your custom bot configuration
}

type CustomEvent struct {
	Data string // just an example of attaching any data with a custom event
}

func main() {
	b := &ExampleBot{
		Bot:     joe.New("example"),
		Channel: "CDEADBEAF", // example reference to a slack channel
	}

	// Register our custom event handler. Joe inspects the function signature to
	// understand that this function should be invoked whenever a CustomEvent
	// is emitted.
	b.Brain.RegisterHandler(b.HandleCustomEvent)

	// For example purposes emit a CustomEvent in a second.
	time.AfterFunc(time.Second, func() {
		b.Brain.Emit(CustomEvent{Data: "Hello World!"})
	})

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

// HandleCustomEvent handles any CustomEvent that is emitted. Joe also supports
// event handlers that return an error or accept a context.Context as first argument.
func (b *ExampleBot) HandleCustomEvent(evt CustomEvent) {
	b.Say(b.Channel, "Received custom event: %v", evt.Data)
}
```

### Integrating with other applications

You may want to integrate your bot with applications such as Github or Gitlab to
trigger a handler or just send a message to Slack. Usually this is done by
providing an HTTP callback to those applications so they can POST data when
there is an event. We already saw in the previous section that is is very easy
to implement custom events so we will use this feature to implement HTTP
integrations as well. Since this is such a dominant use-case we already provide
the [`github.com/go-joe/http-server`][joe-http] module to make it easy for
everybody to write their own custom integrations.

```go
package main

import (
	"errors"
	"github.com/go-joe/http-server"
	"github.com/go-joe/joe"
)

type ExampleBot struct {
	*joe.Bot
}

func main() {
	b := &ExampleBot{Bot: joe.New("example",
		joehttp.Server(":8080"),
	)}

	b.Brain.RegisterHandler(b.HandleHTTP)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func (b *ExampleBot) HandleHTTP(context.Context, joehttp.RequestEvent) error {
	return errors.New("TODO: Add your custom logic here")
}
```

## Available modules

Joe ships with no third-party modules such as Redis integration to avoid pulling
in more dependencies than you actually require. There are however already some
modules that you can use directly to extend the functionality of your bot without
writing too much code yourself.

If you have written a module and want to share it, please add it to this list and
open a pull request.

- Slack Adapter: https://github.com/go-joe/slack-adapter
- Redis Memory: https://github.com/go-joe/redis-memory
- File Memory: https://github.com/go-joe/file-memory
- HTTP Server: https://github.com/go-joe/http-server

## Built With

* [zap](https://github.com/uber-go/zap) - Blazing fast, structured, leveled logging in Go
* [pkg/errors](https://github.com/pkg/errors) - Simple error handling primitives
* [testify](https://github.com/stretchr/testify) - A simple unit test library

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of
conduct and on the process for submitting pull requests to this repository.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available,
see the [tags on this repository][tags]. 

## Authors

- **Friedrich Große** - *Initial work* - [fgrosse](https://github.com/fgrosse)

See also the list of [contributors][contributors] who participated in this project.

## License

This project is licensed under the BSD-3-Clause License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Hubot][hubot] and its great community for the inspiration

[go]: https://golang.org
[hubot]: https://hubot.github.com/
[go-modules]: https://github.com/golang/go/wiki/Modules
[joe-http]: https://github.com/go-joe/http-server
[tags]: https://github.com/go-joe/joe/tags
[contributors]: https://github.com/github.com/go-joe/joe/contributors
