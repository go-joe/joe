package joe

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

type MockAdapter struct {
	mock.Mock
}

func (a *MockAdapter) RegisterAt(b *Brain) {
	a.Called(b)
}

func (a *MockAdapter) Send(text, channel string) error {
	args := a.Called(text, channel)
	return args.Error(0)
}

func (a *MockAdapter) Close() error {
	args := a.Called()
	return args.Error(0)
}
