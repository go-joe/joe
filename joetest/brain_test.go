package joetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestEvent struct{ N int }

func TestBrain(t *testing.T) {
	b := NewBrain(t)

	b.Emit(TestEvent{1})
	b.Emit(TestEvent{2})
	b.Emit(TestEvent{3})
	b.Emit(TestEvent{4})

	b.Finish()

	expectedEvents := []interface{}{
		TestEvent{1},
		TestEvent{2},
		TestEvent{3},
		TestEvent{4},
	}

	actualEvents := b.RecordedEvents()
	assert.Equal(t, expectedEvents, actualEvents)
}
