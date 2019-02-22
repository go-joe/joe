package file

import "go.uber.org/zap"

type Option func(brain *Brain) error

func WithLogger(logger *zap.Logger) Option {
	return func(brain *Brain) error {
		brain.logger = logger
		return nil
	}
}

// IDEA: encrypted brain?
// IDEA: only decrypt keys on demand?
