package joe

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage_Respond(t *testing.T) {
	a := new(MockAdapter)
	msg := Message{adapter: a, Channel: "test"}

	a.On("Send", "Hello world, The Answer is 42", "test").Return(nil)
	msg.Respond("Hello %s, The Answer is %d", "world", 42)
	a.AssertExpectations(t)
}

func TestMessage_RespondE(t *testing.T) {
	a := new(MockAdapter)
	msg := Message{adapter: a, Channel: "test"}

	err := errors.New("a wild issue occurred")
	a.On("Send", "Hello world", "test").Return(err)
	actual := msg.RespondE("Hello world")

	assert.Equal(t, err, actual)
	a.AssertExpectations(t)
}
