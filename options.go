package joe

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	Context        context.Context
	Name           string
	HandlerTimeout time.Duration

	logger  *zap.Logger
	brain *Brain
	adapter Adapter
	errs    []error
}

type EventEmitter interface {
	Emit(event interface{}, callbacks ...func(event))
}

func (c *Config) Logger(name string) *zap.Logger {
	return c.logger.Named(name)
}

func (c *Config) SetMemory(mem Memory) {
	c.brain.memory = mem
}

func (c *Config) SetAdapter(a Adapter) {
	c.adapter = a
}

func (c *Config) RegisterHandler(fun interface{}) {
	c.brain.RegisterHandler(fun)
}

func (c *Config) EventEmitter() EventEmitter {
	return c.brain
}

type Option Module

func WithContext(ctx context.Context) Option {
	return func(conf *Config) error {
		conf.Context = ctx
		return nil
	}
}

func WithHandlerTimeout(timeout time.Duration) Option {
	return func(conf *Config) error {
		conf.HandlerTimeout = timeout
		return nil
	}
}
