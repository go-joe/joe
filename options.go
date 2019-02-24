package joe

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	Context        context.Context
	Name           string
	Logger         *zap.Logger
	Adapter        Adapter
	Memory         Memory
	HandlerTimeout time.Duration

	errs []error
}

func (conf *Config) ApplyOptions(opts []Option) {
	for _, opt := range opts {
		err := opt(conf)
		if err != nil {
			conf.errs = append(conf.errs, err)
		}
	}

	if conf.Adapter == nil {
		conf.Adapter = NewCLIAdapter(conf.Context, conf.Name)
	}

	if conf.Memory == nil {
		conf.Memory = newInMemory()
	}
}

type Option func(*Config) error

func WithContext(ctx context.Context) Option {
	return func(conf *Config) error {
		conf.Context = ctx
		return nil
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(conf *Config) error {
		conf.Logger = logger
		return nil
	}
}

func WithAdapter(adapter Adapter) Option {
	return func(conf *Config) error {
		conf.Adapter = adapter
		return nil
	}
}

func WithMemory(memory Memory) Option {
	return func(conf *Config) error {
		conf.Memory = memory
		return nil
	}
}

func WithHandlerTimeout(timeout time.Duration) Option {
	return func(conf *Config) error {
		conf.HandlerTimeout = timeout
		return nil
	}
}
