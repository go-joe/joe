package redis

import "go.uber.org/zap"

type Option func(brain *Brain) error

func WithLogger(logger *zap.Logger) Option {
	return func(brain *Brain) error {
		brain.logger = logger
		return nil
	}
}

func WithKey(string string) Option {
	return func(brain *Brain) error {
		return nil
	}
}

// TODO: database
// TODO: password
