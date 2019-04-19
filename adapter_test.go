package joe_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/go-joe/joe"
	"github.com/go-joe/joe/joetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func cliTestAdapter(t *testing.T) (a *joe.CLIAdapter, output *bytes.Buffer) {
	logger := zaptest.NewLogger(t)
	a = joe.NewCLIAdapter("test", logger)
	output = new(bytes.Buffer)
	a.Output = output
	a.Author = "TestUser" // ensure tests never depend on external factors such as os.Getenv(…)
	return a, output
}

func TestCLIAdapter_Register(t *testing.T) {
	input := new(bytes.Buffer)
	a, output := cliTestAdapter(t)
	a.Input = ioutil.NopCloser(input)
	brain := joetest.NewBrain(t)
	messages := brain.Events()

	input.WriteString("Hello\n")
	input.WriteString("World\n")

	// Start the Goroutine of the adapter which consumes the input
	a.RegisterAt(brain.Brain)

	msg1 := <-messages
	msg2 := <-messages

	assert.Equal(t, "Hello", msg1.Data.(joe.ReceiveMessageEvent).Text)
	assert.Equal(t, "World", msg2.Data.(joe.ReceiveMessageEvent).Text)

	// Stop the brain to make sure we are done with all callbacks
	brain.Finish()

	// Close the adapter to finish up the test
	assert.NoError(t, a.Close())
	assert.Contains(t, output.String(), "test > ")
}

func TestCLIAdapter_Send(t *testing.T) {
	a, output := cliTestAdapter(t)
	err := a.Send("Hello World", "")
	require.NoError(t, err)
	assert.Equal(t, "Hello World\n", output.String())
}

func TestCLIAdapter_Send_Author(t *testing.T) {
	input := new(bytes.Buffer)
	a, _ := cliTestAdapter(t)
	a.Input = ioutil.NopCloser(input)
	a.Author = "Friedrich"
	brain := joetest.NewBrain(t)
	messages := brain.Events()

	input.WriteString("Test\n")

	// Start the Goroutine of the adapter which consumes the input
	a.RegisterAt(brain.Brain)

	msg := <-messages
	assert.Equal(t, "Friedrich", msg.Data.(joe.ReceiveMessageEvent).AuthorID)

	brain.Finish()
	assert.NoError(t, a.Close())
}

func TestCLIAdapter_Close(t *testing.T) {
	input := new(bytes.Buffer)
	a, output := cliTestAdapter(t)
	a.Input = ioutil.NopCloser(input)
	brain := joe.NewBrain(a.Logger)
	a.RegisterAt(brain)

	err := a.Close()
	require.NoError(t, err)
	assert.Equal(t, "\n", output.String())

	err = a.Close()
	assert.EqualError(t, err, "already closed")
}
