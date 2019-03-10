package joe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)
	brain := NewBrain(logger)
	conf := Config{
		brain:  brain,
		logger: logger,
	}

	assert.Equal(t, brain, conf.EventEmitter())
	assert.NotNil(t, logger, conf.Logger("test"))

	adapter := new(MockAdapter)
	conf.SetAdapter(adapter)
	assert.Equal(t, adapter, conf.adapter)

	mem := newInMemory()
	conf.SetMemory(mem)
	assert.Equal(t, mem, brain.memory)

	conf.RegisterHandler(func(InitEvent) {})
}

func TestWithContext(t *testing.T) {
	var conf Config
	mod := WithContext(ctx)
	err := mod(&conf)
	assert.NoError(t, err)
	assert.Equal(t, ctx, conf.Context)
}

func TestWithHandlerTimeout(t *testing.T) {
	var conf Config
	mod := WithHandlerTimeout(42 * time.Millisecond)
	err := mod(&conf)
	assert.NoError(t, err)
	assert.Equal(t, 42*time.Millisecond, conf.HandlerTimeout)
}
