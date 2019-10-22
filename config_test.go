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
	store := NewStorage(logger)
	conf := Config{
		brain:  brain,
		store:  store,
		logger: logger,
	}

	assert.Equal(t, brain, conf.EventEmitter())
	assert.NotNil(t, logger, conf.Logger("test"))

	adapter := new(MockAdapter)
	conf.SetAdapter(adapter)
	assert.Equal(t, adapter, conf.adapter)

	mem := newInMemory()
	conf.SetMemory(mem)
	assert.Equal(t, mem, store.memory)

	enc := jsonEncoder{}
	conf.SetMemoryEncoder(enc)
	assert.Equal(t, enc, store.encoder)

	conf.RegisterHandler(func(InitEvent) {})
}

func TestWithContext(t *testing.T) {
	var conf Config
	mod := WithContext(ctx)
	err := mod.Apply(&conf)
	assert.NoError(t, err)
	assert.Equal(t, ctx, conf.Context)
}

func TestWithHandlerTimeout(t *testing.T) {
	var conf Config
	mod := WithHandlerTimeout(42 * time.Millisecond)
	err := mod.Apply(&conf)
	assert.NoError(t, err)
	assert.Equal(t, 42*time.Millisecond, conf.HandlerTimeout)
}

// TestBot_Logger simply tests that the zap logger configuration in newLogger()
// doesn't panic.
func TestBot_Logger(t *testing.T) {
	newLogger(nil)
}
