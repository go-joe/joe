+++
title = "Working with Web Applications"
slug = "integration"
weight = 5
pre = "<b>e) </b>"
+++

{{% notice note %}}
More information on how to consume HTTP callbacks can now be found in the [**Events recipe**](/recipes/events/#chaining-events)
{{% /notice %}}

You may want to integrate your bot with applications such as GitHub or GitLab to
trigger a handler or just send a message to Slack. Usually this is done by
providing an HTTP callback to those applications so they can POST data when
there is an event.

We already saw in the previous section that is is very easy to implement custom
events so we will use this feature to implement HTTP integrations as well. Since
this is such a dominant use-case we already provide the
[`github.com/go-joe/http-server`][joe-http] module to make it easy for everybody
to write their own custom integrations.

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

[joe-http]: https://github.com/go-joe/http-server
