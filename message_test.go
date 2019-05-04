package joe

import (
	"errors"
	"testing"

	"github.com/go-joe/joe/reactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMessage_Respond(t *testing.T) {
	a := new(MockAdapter)
	msg := Message{ReceiveMessageEvent: ReceiveMessageEvent{Adapter: a, Channel: "test"}}

	a.On("Send", "Hello world, The Answer is 42", "test").Return(nil)
	msg.Respond("Hello %s, The Answer is %d", "world", 42)
	a.AssertExpectations(t)
}

func TestMessage_RespondE(t *testing.T) {
	a := new(MockAdapter)
	msg := Message{ReceiveMessageEvent: ReceiveMessageEvent{Adapter: a, Channel: "test"}}

	err := errors.New("a wild issue occurred")
	a.On("Send", "Hello world", "test").Return(err)
	actual := msg.RespondE("Hello world")

	assert.Equal(t, err, actual)
	a.AssertExpectations(t)
}

func TestMessage_React_NotImplemented(t *testing.T) {
	a := new(MockAdapter)
	msg := Message{ReceiveMessageEvent: ReceiveMessageEvent{Adapter: a}}

	err := msg.React(reactions.Thumbsup)
	assert.Equal(t, ErrNotImplemented, err)
	a.AssertExpectations(t)
}

func TestMessage_React(t *testing.T) {
	a := new(ExtendedMockAdapter)
	msg := Message{ReceiveMessageEvent: ReceiveMessageEvent{Adapter: a}}

	err := errors.New("this clearly failed")
	a.On("React", reactions.Thumbsup, msg).Return(err)
	actual := msg.React(reactions.Thumbsup)

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

type ExtendedMockAdapter struct {
	MockAdapter
}

func (a *ExtendedMockAdapter) React(r reactions.Reaction, msg Message) error {
	args := a.Called(r, msg)
	return args.Error(0)
}
