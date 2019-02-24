package slack

import (
	"go.uber.org/zap"
)

type Option func(*Config) error

func WithLogger(logger *zap.Logger) Option {
	return func(conf *Config) error {
		conf.Logger = logger
		return nil
	}
}

func WithDebug(debug bool) Option {
	return func(conf *Config) error {
		conf.Debug = debug
		return nil
	}
}

// IDEA: encrypted brain?
// IDEA: only decrypt keys on demand?
