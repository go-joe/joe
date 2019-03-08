package joe

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func cliTestAdapter(t *testing.T) (a *CLIAdapter, output *bytes.Buffer) {
	logger := zaptest.NewLogger(t)
	a = NewCLIAdapter("test", logger)
	output = new(bytes.Buffer)
	a.Output = output
	return a, output
}

func TestCLIAdapter_Register(t *testing.T) {
	input := new(bytes.Buffer)
	a, output := cliTestAdapter(t)
	a.Input = ioutil.NopCloser(input)
	brain := NewBrain(a.Logger)

	input.WriteString("Hello\n")
	input.WriteString("World\n")

	messages := make(chan ReceiveMessageEvent, 2)
	brain.RegisterHandler(func(msg ReceiveMessageEvent) {
		messages <- msg
	})

	brain.connectAdapter(a)

	ctx, stopBrain := context.WithCancel(context.Background())
	brainStopped := make(chan bool)
	go func() {
		brain.HandleEvents(ctx)
		brainStopped <- true
	}()

	msg1 := <-messages
	msg2 := <-messages

	assert.Equal(t, "Hello", msg1.Text)
	assert.Equal(t, "World", msg2.Text)

	// Stop the brain to make sure we are done with all callbacks
	stopBrain()
	<-brainStopped

	// Close the adapter to finish up the test
	assert.NoError(t, a.Close())

	expectedOutput := strings.Join([]string{
		"test > ", // Hello
		"test > ", // World
		"test > ", // <ctrl>+c
		"\n",
	}, "")
	assert.Equal(t, expectedOutput, output.String())
}

func TestCLIAdapter_Send(t *testing.T) {
	a, output := cliTestAdapter(t)
	err := a.Send("Hello World", "")
	require.NoError(t, err)
	assert.Equal(t, "Hello World\n", output.String())
}

func TestCLIAdapter_Close(t *testing.T) {
	input := new(bytes.Buffer)
	a, output := cliTestAdapter(t)
	a.Input = ioutil.NopCloser(input)
	brain := NewBrain(a.Logger)
	brain.connectAdapter(a)

	err := a.Close()
	require.NoError(t, err)
	assert.Equal(t, "\n", output.String())

	err = a.Close()
	assert.EqualError(t, err, "already closed")
}
