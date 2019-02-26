package joe

import (
	"context"
	"fmt"
)

func ExampleBrain_RegisterHandler() {
	done := make(chan bool)
	type CustomEvent struct{ Test bool }

	b := NewBrain(nil)
	b.RegisterHandler(func(event CustomEvent) {
		fmt.Printf("Received custom event: %+v\n", event)
		done <- true
	})

	ctx, cancel := context.WithCancel(context.Background())
	b.Emit(CustomEvent{Test: true})
	go b.HandleEvents(ctx)

	<-done
	cancel()
	// Output: Received custom event: {Test:true}
}
