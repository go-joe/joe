+++
title = "The Event System"
slug = "events"
weight = 3
+++

Joe is powered by an asynchronous event system which is also used internally when
you register your message handlers using the [`Bot.Respond(…)`][1] function.
What happens is that the chat Adapter emits a [`joe.ReceiveMessageEvent`][2] for each
incoming message. The handler function that you registered is executed if the bot
sees such an event with a message text that matches your regular expression.

```go
func (b *Bot) RespondRegex(expr string, fun func(Message) error) {
	// other code omitted for brevity …

	b.Brain.RegisterHandler(func(ctx context.Context, evt ReceiveMessageEvent) error {
		matches := regex.FindStringSubmatch(evt.Text)
		if len(matches) == 0 {
			return nil
		}

		// If the event text matches our regular expression we can already mark
		// the event context as done so the Brain does not run any other handlers
		// that might match the received message.
		FinishEventContent(ctx)

		return fun(Message{
			ID:       evt.ID,
			Text:     evt.Text,
			// and other many fields
		})
	})
}
```

This means that you could also register message handlers directly on the
`joe.ReceiveMessageEvent` yourself if you wanted to (e.g. if you want to get
notified for each incoming message).

### The Brain

This event system is implemented in the [`joe.Brain`][4]. When it sees a new event it
finds all registered event handlers for the event type and then executes them
all in the same sequence in which they have been registered.

By default _all_ matching handlers are executed but you can prevent other
handlers from being executed after your handler, by calling the
[`joe.FinishEventContent()`][3] function. As you can see in the code snippet above,
this is also what happens automatically to message handlers you register via `Bot.Respond(…)`.

The Brain itself also emits two events to signal when it is starting up and when it
is shutting down:

- `joe.InitEvent`
- `joe.ShutdownEvent`

You can use those events both in unit tests as well as your own logic to hook into
the lifecycle of the bot.

### Chaining events

The event system is also useful for other kinds of events. For instance, as you
can see in the next recipes, this is how [cron jobs](/recipes/cron) are
implemented in Joe. Generally speaking, there are _sources_ that can trigger
events and there are _handlers_ that get executed when the matching event is
emitted. Since handlers can also be event sources, this means you can chain
events asynchronously.

For example we can setup a handler that should be executed for each incoming HTTP
request. It should check if the request came from GitLab and if so, decode the
request body it into another event type:

```go
import (
	"fmt"

	"encoding/json"
	joehttp "github.com/go-joe/http-server"
)

func (b *Bot) HTTPCallback(req joehttp.RequestEvent) error {
	if req.Header.Get("X-Gitlab-Event") == "" {
		return nil
	}

	var event GitLabEvent
	err := json.Unmarshal(req.Body, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal gitlab event as JSON: %w", err)
	}

	b.Brain.Emit(event)
	return nil
}
``` 

Now we can define another handler that will be executed on the `GitLabEvent` type:

```go
func (b *Bot) GitLabCallback(event GitLabEvent) error {
	b.Logger.Info("Received gitlab event",
		zap.String("event_type", event.EventType),
		zap.String("object_kind", event.ObjectKind),
		zap.String("action", event.ObjectAttributes.Action),
		zap.String("project", event.Project.PathWithNamespace),
		zap.String("title", event.ObjectAttributes.Title),
		zap.String("url", event.ObjectAttributes.URL),
	)

	switch event.EventType {
	case "merge_request":
		return b.HandleMergeRequestEvent(event)

	case "note":
		return b.HandleGitlabNoteEvent(event)

	default:
		b.Logger.Info("Unknown event from gitlab", zap.String("object_kind", event.ObjectKind))
		return nil
	}
}
```

Finally to make this all work together, we need to register the two handlers
when we setup the bot:

```go
func New(conf Config) *Bot {
	b := &Bot{
		Bot: joe.New("joe", conf.Modules()...),
	}

	// Define any custom event and message handlers here
	b.Brain.RegisterHandler(b.HTTPCallback)
	b.Brain.RegisterHandler(b.GitLabCallback)
	
	return b
}
```

If you want to learn more about how the Brain works internally, start by looking
at [the GoDoc][4] and then [the code itself][5].

Happy event hacking :robot:.

[1]: https://godoc.org/github.com/go-joe/joe#Bot.Respond
[2]: https://godoc.org/github.com/go-joe/joe#ReceiveMessageEvent
[3]: https://godoc.org/github.com/go-joe/joe#FinishEventContent
[4]: https://godoc.org/github.com/go-joe/joe#Brain
[5]: https://github.com/go-joe/joe/blob/master/brain.go
