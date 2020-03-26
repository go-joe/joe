<h1 align="center">Joe Bot :robot:</h1>
<p align="center">A general-purpose bot library inspired by Hubot but written in Go.</p>
<p align="center">
    <a href="https://joe-bot.net"><img src="https://img.shields.io/badge/website-joe--bot.net-brightgreen"></a>
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

Joe is a software library that is packaged as [Go module][go-modules]. You can get it via:

```
go get github.com/go-joe/joe
```

### Example usage

**You can find all code examples, more explanation and complete recipes at https://joe-bot.net**

Each bot consists of a chat _Adapter_ (e.g. to integrate with Slack), a _Memory_
implementation to remember key-value data (e.g. using Redis) and a _Brain_ which
routes new messages or custom events (e.g. receiving an HTTP call) to the
corresponding registered _handler_ functions.

By default `joe.New(…)` uses the CLI adapter which makes the bot read messages
from stdin and respond on stdout. Additionally the bot will store key value
data in-memory which means it will forget anything you told it when it is restarted.
This default setup is useful for local development without any dependencies but
you will quickly want to add other _Modules_ to extend the bots capabilities.

The following example connects the Bot with a Slack workspace and stores
key-value data in Redis. To allow the message handlers to access the memory we
define them as functions on a custom `ExampleBot`type which embeds the `joe.Bot`.

[embedmd]:# (_examples/02_useful/main.go)
```go
package main

import (
	"fmt"

	"github.com/go-joe/joe"
	"github.com/go-joe/redis-memory"
	"github.com/go-joe/slack-adapter"
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
		return fmt.Errorf("failed to retrieve key %q from brain: %w", key, err)
	}

	if ok {
		msg.Respond("%s is %s", key, value)
	} else {
		msg.Respond("I do not remember %q", key)
	}

	return nil
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
- Mattermost Adapter: https://github.com/dwmunster/joe-mattermost-adapter
- VK Adapter: https://github.com/tdakkota/joe-vk-adapter

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
* [multierr](https://github.com/uber-go/multierr) - Package multierr allows combining one or more errors together 
* [testify](https://github.com/stretchr/testify) - A simple unit test library

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of
conduct and on the process for submitting pull requests to this repository.

## Versioning

**THIS SOFTWARE IS STILL IN ALPHA AND THERE ARE NO GUARANTEES REGARDING API STABILITY YET.**

All significant (e.g. breaking) changes are documented in the [CHANGELOG.md](CHANGELOG.md).

After the v1.0 release we plan to use [SemVer](http://semver.org/) for versioning.
For the versions available, see the [tags on this repository][tags]. 

## Authors

- **Friedrich Große** - *Initial work* - [fgrosse](https://github.com/fgrosse)

See also the list of [contributors][contributors] who participated in this project.

## License

This project is licensed under the BSD-3-Clause License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Hubot][hubot] and its great community for the inspiration
- [embedmd][campoy-embedmd] for a cool tool to embed source code in markdown files

[go]: https://golang.org
[hubot]: https://hubot.github.com/
[go-modules]: https://github.com/golang/go/wiki/Modules
[joe-http]: https://github.com/go-joe/http-server
[tags]: https://github.com/go-joe/joe/tags
[contributors]: https://github.com/go-joe/joe/contributors
[campoy-embedmd]: https://github.com/campoy/embedmd
