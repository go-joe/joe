+++
title = "Quick Start"
slug = "quick"
weight = 1
pre = "<b>1. </b>"
+++

## Installation

Joe is a software library that is packaged as [Go module][go-modules]. You can get it via:

```sh
go get github.com/go-joe/joe
```

### Your First Bot

{{< include "/basics/01_minimal.md" >}}

### Run it

To run the code above, save it as `main.go` and then execute it via `go run main.go`. By default Joe uses
a CLI adapter which makes the bot read messages from stdin and respond on stdout.

### Next Steps

Please refer to the [**Basic Usage**](/basics) section to learn how to write a
full Joe Bot, using the [adapter](/modules/#chat-adapters) of your choice (e.g. Slack). If you want to dive
right in and want to know what modules are currently provided by the community,
then have a look at the [**Available Modules**](/modules) section. Last but not
least, you can find more instructions and best practices in the [**Recipes**](/recipes) section. 

Happy hacking :robot:

[go-modules]: https://github.com/golang/go/wiki/Modules
