package joe

import (
	"context"

	"go.uber.org/zap"
)

// Config is the configuration of a Bot that can be used or changed during setup
// in a Module. Some configuration settings such as the Logger are read only can
// only be accessed via the corresponding getter function of the Config.
type Config struct {
	Context context.Context
	Name    string

	logger  *zap.Logger
	brain   *Brain
	adapter Adapter
	errs    []error
}

// The EventEmitter can be used by a Module by calling Config.EventEmitter().
// Events are emitted asynchronously so every call to Emit is non-blocking.
type EventEmitter interface {
	Emit(event interface{}, callbacks ...func(Event))
}

// EventEmitter returns the EventEmitter that can be used to send events to the
// Bot and other modules.
func (c *Config) EventEmitter() EventEmitter {
	return c.brain
}

// Logger returns a new named logger.
func (c *Config) Logger(name string) *zap.Logger {
	return c.logger.Named(name)
}

// SetMemory can be used to change the Memory implementation of the bots Brain.
func (c *Config) SetMemory(mem Memory) {
	c.brain.memory = mem
}

// SetAdapter can be used to change the Adapter implementation of the Bot.
func (c *Config) SetAdapter(a Adapter) {
	c.adapter = a
}

// RegisterHandler can be used to register an event handler in a Module.
func (c *Config) RegisterHandler(fun interface{}) {
	c.brain.RegisterHandler(fun)
}

// WithContext is an option to replace the default context of a bot.
func WithContext(ctx context.Context) Module {
	return func(conf *Config) error {
		conf.Context = ctx
		return nil
	}
}
