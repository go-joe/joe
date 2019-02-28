package joe

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func cliTestAdapter(t *testing.T) (a *CLIAdapter, output *bytes.Buffer) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	a = NewCLIAdapter(ctx, "test", logger)
	output = new(bytes.Buffer)
	a.Output = output
	return a, output
}

func TestCLIAdapter_Register(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())

	input := new(bytes.Buffer)
	a, output := cliTestAdapter(t)
	a.Input = input
	brain := NewBrain(a.Logger)

	input.WriteString("Hello\n")
	input.WriteString("World\n")

	messages := make(chan ReceiveMessageEvent, 2)
	brain.RegisterHandler(func(msg ReceiveMessageEvent) {
		messages <- msg
	})

	a.Register(brain)
	go brain.HandleEvents(ctx)

	msg1 := <-messages
	msg2 := <-messages

	assert.Equal(t, "Hello", msg1.Text)
	assert.Equal(t, "World", msg2.Text)
	assert.Equal(t, "test > test > test > ", output.String())

	done()
}

func TestCLIAdapter_Send(t *testing.T) {
	a, output := cliTestAdapter(t)
	err := a.Send("Hello World", "")
	require.NoError(t, err)
	assert.Equal(t, "Hello World\n", output.String())
}

func TestCLIAdapter_Close(t *testing.T) {
	a, output := cliTestAdapter(t)
	err := a.Close()
	require.NoError(t, err)
	assert.Equal(t, "\n", output.String())
}
