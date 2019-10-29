+++
title = "Cron Jobs"
slug = "cron"
weight = 4
+++

The [Cron module][module] allows you to run arbitrary functions or [emit events](/recipes/events)
on a schedule using cron expressions or a specific time interval.

For instance you might want to trigger a function that should be executed every
day at midnight.

```go
import "github.com/go-joe/cron"

func main() {
	b := joe.New("example-bot",
		cron.ScheduleEvent("0 0 * * *"),
	)
	
	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func AtMidnight(cron.Event) {
	// do something spooky 👻
}
```

The `cron.ScheduleEvent` will emit a `cron.Event` at `"0 0 * * *"` which is the
[cron expression][cron] for "every day at midnight".

If you find cron expressions hard to read, you can also use the `cron.ScheduleEventEvery(…)`
function which accepts a `time.Duration` as the first argument.

```go
import (
	"time"
	"github.com/go-joe/cron"
)

func main() {
	b := joe.New("example-bot",
		cron.ScheduleEventEvery(time.Hour),
	)
	
	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func EveryHour(cron.Event) {
	// do something funky 🤘
}
```

If you have multiple cron jobs running (e.g. the one at midnight and the hourly one)
you now have the problem that your two registered functions get executed for _every_
`cron.Event`. Instead what you actually want is that `EveryHour` is executed
independently of `AtMidnight`. This can be fixed by emitting your own custom event types:

```go
type MidnightEvent struct{}

type HourEvent struct {}

func main() {
	b := joe.New("example-bot",
		cron.ScheduleEvent("0 0 * * *", MidnightEvent{}),
		cron.ScheduleEventEvery(time.Hour, HourEvent{}),
	)
	
	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func AtMidnight(MidnightEvent) {
	// 👻👻👻
}

func EveryHour(HourEvent) {
	// ⏰⏰⏰
}
```

Emitting your own events can also be useful if you want to trigger existing handlers
in more than one way (e.g. directly and via cron):

```go
type DoStuffEvent struct {}

func main() {
	b := joe.New("example-bot",
		cron.ScheduleEventEvery(time.Hour, DoStuffEvent{}),
	)
	
	b.Respond("do it", func(joe.Message) error {
		b.Brain.Emit(DoStuffEvent{})
	})
	
	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func DoStuff(DoStuffEvent) {
	// Do this every hour and when the user asks us to
}
```

In practice you will likely want to execute functions directly without having to
create those extra types. This can be done with the `cron.ScheduleFunc(…)` and
`cron.ScheduleFuncEvery(…)` functions which work like the functions we saw early
just with closures instead of event types:

```go
package main

import (
	"time"
	"github.com/go-joe/joe"
	"github.com/go-joe/cron"
)

type MyEvent struct {}

func main() {
	b := joe.New("example-bot",
		// emit a cron.Event once every day at midnight
		cron.ScheduleEvent("0 0 * * *"),
		
		// emit your own custom event every day at 09:00
		cron.ScheduleEvent("0 9 * * *", MyEvent{}), 
		
		// cron expressions can be hard to read and might be overkill
		cron.ScheduleEventEvery(time.Hour, MyEvent{}), 
		
		// sometimes its easier to use a function
		cron.ScheduleFunc("0 9 * * *", func() { /* TODO */ }), 
		
		// functions can also be scheduled on simple intervals
		cron.ScheduleFuncEvery(5*time.Minute, func() { /* TODO */ }),
	)
	
	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}
```

[module]: https://github.com/go-joe/cron
[cron]: https://en.wikipedia.org/wiki/Cron#Overview
