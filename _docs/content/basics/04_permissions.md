+++
title = "Checking User Permissions"
slug = "permissions"
weight = 4
pre = "<b>d) </b>"
+++

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

[embedmd]:# (../../../_examples/04_auth/main.go)
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
