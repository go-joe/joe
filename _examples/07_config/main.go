package main

import (
	"log"
)

func main() {
	b, err := New2(Config{
		SlackToken: "xoxb-1452345…",
		HTTPListen: ":80",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}
