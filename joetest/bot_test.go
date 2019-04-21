package joetest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBot(t *testing.T) {
	b := NewBot(t)

	var seenEvents []TestEvent
	b.Brain.RegisterHandler(func(evt TestEvent) {
		seenEvents = append(seenEvents, evt)
	})

	b.Start()
	assert.Equal(t, "test > ", b.ReadOutput())

	b.EmitSync(TestEvent{N: 123})
	b.Stop()

	assert.Equal(t, []TestEvent{{N: 123}}, seenEvents)
}

func TestBotEmitSyncTimeout(t *testing.T) {
	mock := new(mockT)
	b := NewBot(mock)
	b.Timeout = time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b.Brain.RegisterHandler(func(evt TestEvent) {
		<-ctx.Done()
	})

	b.Start()
	b.EmitSync(TestEvent{})
	b.Stop()

	require.Len(t, mock.Errors, 2)
	assert.True(t, mock.failed)
	assert.True(t, mock.fatal)
	assert.Equal(t, "EmitSync timed out", mock.Errors[0])
	assert.Equal(t, "Stop timed out", mock.Errors[1])
}
