+++
title = "Advanced Slack Features"
slug = "slack"
weight = 2
+++

One of the first questions of new users of Joe is, whether feature _X_, which
they know from Slack, is supported in Joe as well. Usually the answer is, that
we want to keep the generic `Adapter` interface minimal and thus we cannot add
direct support for this feature since it would make it very hard to port it over
to other chat adapters (e.g. to IRC).

Luckily however there is a pattern that you can use which allows you to use all
features of the Chat adapter of your choice, without requiring the Joe library 
to know about them.

First you need to define a custom type for your bot. This is a useful pattern in
any case since it simplifies writing your handlers when you need access to other
Joe features such as memory persistence or authentication.

```go
package main

import (
	"github.com/go-joe/joe"
	"github.com/nlopes/slack"
)

// Bot is your own custom bot type. It has access to all features of the slack
// API in its handler functions.
type Bot struct {
	*joe.Bot // Anonymously embed the joe.Bot type so we can use its functions easily.
	Slack *slack.Client 
}
``` 

To create a new instance of your bot you probably want to implement a `New(…)`
function. Details on how to configure your bot when you have multiple parameters
can be found in the [Bot Configuration](/recipes/config) recipe.

```go
import (
	"github.com/go-joe/joe"
	joeSlack "github.com/go-joe/slack-adapter/v2"
	"github.com/nlopes/slack"
)

func New(slackToken string) *Bot {
	b := &Bot{
		Bot:   joe.New("joe", joeSlack.Adapter(slackToken)),
		Slack: slack.New(slackToken),
	}
	
	b.Respond("do stuff", b.DoStuffCommand)
	
	// other setup may happen here as well     
	
	return b
}
```

Now when you have a handler that should use a slack specific feature you can
define it as function on your own `Bot` type and use the `Slack` field
to access the client directly.

```go
func (b *Bot) DoStuffCommand(msg joe.Message) error {
	if b.Slack == nil {
		// In case this command does not even make sense without your custom
		// functionality you may want to return early in the command.
		// Having such a check is only useful if you actually create the Bot such
		// that users can create a new instance even if they do not provide a slack token. 
		return msg.RespondE("Cannot do stuff because Slack integration is not enabled")
	}

	// Access to specific functionality is accessible via an extra field on the Bot.
	// This example uses a specific feature of Slack to style the message using
	// multiple message blocks. Of course in practice you have access to all features
	// exposed to you via the Slack Go client.
	var blocks []slack.Block
	blocks = append(blocks, slack.NewSectionBlock("Foo!", nil, nil))
	blocks = append(blocks, slack.NewDividerBlock())
	blocks = append(blocks, b.createMessageBlocks()...)
	blocks = append(blocks, slack.NewDividerBlock())
	
	_, _, err := b.Slack.PostMessageContext(ctx, channel,
	    slack.MsgOptionBlocks(blocks...),
	    slack.MsgOptionPostMessageParameters(
	        slack.PostMessageParameters{
	            LinkNames: 1,
	            Parse:     "full",
	            AsUser:    true,
	        },
	    ),
	)
	
	// You can still use all regular adapter features easily.
	msg.Respond("OK")
	return nil
}
```

Of course you can use the same pattern to access specific features of other chat
adapters as well. 

In case you need to close your chat adapter client explicitly when the bot is
shutting down, you can register a shutdown event handler function like this:

```go

func New(slackToken string) *Bot {
	b := &Bot{
		Bot:   joe.New("joe", joeSlack.Adapter(slackToken)),
		Slack: slack.New(slackToken),
	}
	
	// other setup tasks are usually here
	
	b.Brain.RegisterHandler(b.Shutdown)     

	return b
}



func (b *Bot) Shutdown(joe.ShutdownEvent) {
	// TODO: implement your cleanup logic
}
```
