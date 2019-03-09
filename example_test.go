package joe

import "fmt"

func ExampleBrain_RegisterHandler() {
	done := make(chan bool) // just to cleanly shutdown when we processed the event

	type CustomEvent struct{ Test bool }

	b := NewBrain(nil)
	b.RegisterHandler(func(event CustomEvent) {
		fmt.Printf("Received custom event: %+v\n", event)
		done <- true
	})

	go b.HandleEvents()
	b.Emit(CustomEvent{Test: true})

	<-done
	b.Shutdown()

	// Output: Received custom event: {Test:true}
}
