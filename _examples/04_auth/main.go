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
