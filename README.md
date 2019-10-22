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

## Getting Started

**THIS SOFTWARE IS STILL IN ALPHA AND THERE ARE NO GUARANTEES REGARDING API STABILITY YET.**

All significant (e.g. breaking) changes are documented in the [CHANGELOG.md](CHANGELOG.md).

Joe is packaged using the new [Go modules][go-modules]. You can get joe via:

```
go get github.com/go-joe/joe
```

### Minimal example

The simplest chat bot listens for messages on a chat _Adapter_ and then executes
a _Handler_ function if it sees a message directed to the bot that matches a given pattern.

For example a bot that responds to a message "ping" with the answer "PONG" looks like this:

[embedmd]:# (_examples/01_minimal/main.go)
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
from stdin and respond on stdout. Additionally the bot will store key value
data in-memory which means it will forget anything you told it when it is restarted.
This default setup is useful for local development without any dependencies but
you will quickly want to add other _Modules_ to extend the bots capabilities.

For instance we can extend the previous example to connect the Bot with a Slack
workspace and store key-value data in Redis. To allow the message handlers to
access the memory we define them as functions on a custom `ExampleBot`type which
embeds the `joe.Bot`.

[embedmd]:# (_examples/02_useful/main.go)
```go
package main

import (
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
	return b.Store.Set(key, value)
}

func (b *ExampleBot) WhatIs(msg joe.Message) error {
	key := msg.Matches[0]
	var value string
	ok, err := b.Store.Get(key, &value)
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

[embedmd]:# (_examples/03_custom_events/main.go)
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

### Granting and checking user permissions

Joe supports a simple way to manage user permissions. For instance you may want
to define a message handler that will run an operation which only admins should
be allowed to trigger.

To implement this, joe has a concept of permission scopes. A scope is a string
which is _granted_ to a specific user ID so you can later check if the author of
the event you are handling (e.g. a message from Slack) has this scope or any
scope that _contains_ it.

Scopes are interpreted in a hierarchical way where scope _A_ can contain scope
_B_ if _A_ is a prefix to _B_. For example, you can check if a user is allowed
to read or write from the "Example" API by checking the `api.example.read` or
`api.example.write` scope. When you grant the scope to a user you can now either
decide only to grant the very specific `api.example.read` scope which means the
user will not have write permissions or you can allow people write-only access
via the `api.example.write` scope.

Alternatively you can also grant any access to the Example API via `api.example`
which includes both the read and write scope beneath it. If you want you
could also allow even more general access to everything in the api via the
`api` scope.

Scopes can be granted statically in code or dynamically in a handler like this:

[embedmd]:# (_examples/04_auth/main.go)
```go
package main

import "github.com/go-joe/joe"

type ExampleBot struct {
	*joe.Bot
}

func main() {
	b := &ExampleBot{
		Bot: joe.New("HAL"),
	}

	// If you know the user ID in advance you may hard-code it at startup.
	b.Auth.Grant("api.example", "DAVE")

	// An example of a message handler that checks permissions.
	b.Respond("open the pod bay doors", b.OpenPodBayDoors)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func (b *ExampleBot) OpenPodBayDoors(msg joe.Message) error {
	err := b.Auth.CheckPermission("api.example.admin", msg.AuthorID)
	if err != nil {
		return msg.RespondE("I'm sorry Dave, I'm afraid I can't do that")
	}

	return msg.RespondE("OK")
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

[embedmd]:# (_examples/05_http/main.go)
```go
package main

import (
	"context"
	"errors"

	joehttp "github.com/go-joe/http-server"
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

### Chat Adapters

- Slack Adapter: https://github.com/go-joe/slack-adapter
- Rocket.Chat Adapter: https://github.com/dwmunster/rocket-adapter
- Telegram Adapter: https://github.com/robertgzr/joe-telegram-adapter
- IRC Adapter: https://github.com/akrennmair/joe-irc-adapter

### Memory Modules

- Redis Memory: https://github.com/go-joe/redis-memory
- File Memory: https://github.com/go-joe/file-memory
- Bolt Memory: https://github.com/robertgzr/joe-bolt-memory
- Sqlite Memory: https://github.com/warmans/sqlite-memory

### Other Modules

- HTTP Server: https://github.com/go-joe/http-server
- Cron Jobs: https://github.com/go-joe/cron

## Built With

* [zap](https://github.com/uber-go/zap) - Blazing fast, structured, leveled logging in Go
* [pkg/errors](https://github.com/pkg/errors) - Simple error handling primitives
* [testify](https://github.com/stretchr/testify) - A simple unit test library

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of
conduct and on the process for submitting pull requests to this repository.

## Versioning

**THIS SOFTWARE IS STILL IN ALPHA AND THERE ARE NO GUARANTEES REGARDING API STABILITY YET.**

After the v1.0 release we plan to use [SemVer](http://semver.org/) for versioning.
For the versions available, see the [tags on this repository][tags]. 

## Authors

- **Friedrich Große** - *Initial work* - [fgrosse](https://github.com/fgrosse)

See also the list of [contributors][contributors] who participated in this project.

## License

This project is licensed under the BSD-3-Clause License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Hubot][hubot] and its great community for the inspiration
- [embedmd][embedmd] for a cool tool to embed source code in markdown files

[go]: https://golang.org
[hubot]: https://hubot.github.com/
[go-modules]: https://github.com/golang/go/wiki/Modules
[joe-http]: https://github.com/go-joe/http-server
[tags]: https://github.com/go-joe/joe/tags
[contributors]: https://github.com/go-joe/joe/contributors
[embedmd]: https://github.com/campoy/embedmd
