+++
title = "Bot Configuration"
slug = "config"
weight = 1
+++

After you created your first bot you will likely have some configuration you
want to pass to it when you set it up. Sometimes the available configuration
might even determine what modules the bot should use. For instance if the user
passed a slack token, you can use the Slack chat adapter. Otherwise you might
want to fallback to the CLI adapter so you can easily develop your bot locally.

This tutorial shows a pattern that allows you to implement such a use case.

### The Configuration type 

First we need a structure that can hold all our configurable parameters. For
this tutorial we have the slack token, an HTTP listen address as well as an
address to Redis. Note that all fields are optional since the bot can fallback
to defaults (CLI instead of slack, im-memory instead of Redis) or disable a
feature all together (e.g. no HTTP server).

[embedmd]:# (../../../_examples/07_config/config.go /\/\/ Config holds all/ /return modules\n\}/)
```go
// Config holds all parameters to setup a new chat bot.
type Config struct {
	SlackToken string // slack token, if empty we fallback to the CLI
	HTTPListen string // optional HTTP listen address to receive callbacks
	RedisAddr  string // optional address to store keys in Redis
}

// Modules creates a list of joe.Modules that can be used with this configuration.
func (conf Config) Modules() []joe.Module {
	var modules []joe.Module

	if conf.SlackToken != "" {
		modules = append(modules, slack.Adapter(conf.SlackToken))
	}

	if conf.HTTPListen != "" {
		modules = append(modules, joehttp.Server(conf.HTTPListen))
	}

	if conf.RedisAddr != "" {
		modules = append(modules, redis.Memory(conf.RedisAddr))
	}

	return modules
}
```

We also want to define our own Bot type on which we can define our handlers. To
create a new instance we will also provide a `New(…)` function which accepts the
previously defined configuration type.

[embedmd]:# (../../../_examples/07_config/bot.go /type Bot/ /return b\n\}/)
```go
type Bot struct {
	*joe.Bot        // Anonymously embed the joe.Bot type so we can use its functions easily.
	conf     Config // You can keep other fields here as well.
}

func New(conf Config) *Bot {
	b := &Bot{
		Bot: joe.New("joe", conf.Modules()...),
	}

	// Define any custom event and message handlers here
	b.Brain.RegisterHandler(b.GitHubCallback)
	b.Respond("do stuff", b.DoStuffCommand)

	return b
}
```

From here on you can extend the `New` function to do other setup as well such as
connecting to third-party APIs or setting up cron jobs. If you want to enforce
the existence of some configuration parameters or you generally want to validate
the passed parameters you can do this via a new `Config.Validate()` function that
is called before creating a new Bot:

[embedmd]:# (../../../_examples/07_config/config.go /func \(conf Config\) Validate/ $)
```go
func (conf Config) Validate() error {
	if conf.HTTPListen == "" {
		return errors.New("missing HTTP listen address")
	}
	return nil
}
```

[embedmd]:# (../../../_examples/07_config/bot.go /func New2/ /return b, nil\n\}/)
```go
func New2(conf Config) (*Bot, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	b := &Bot{
		Bot: joe.New("joe", conf.Modules()...),
	}

	// Define any custom event and message handlers here
	b.Brain.RegisterHandler(b.GitHubCallback)
	b.Respond("do stuff", b.DoStuffCommand)

	return b, nil
}
```
