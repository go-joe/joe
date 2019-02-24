package redis

import "go.uber.org/zap"

type Option func(*Config) error

func WithConfig(newConf Config) Option {
	return func(oldConf *Config) error {
		oldConf.Addr = newConf.Addr
		oldConf.Key = newConf.Key
		oldConf.Password = newConf.Password
		oldConf.DB = newConf.DB
		oldConf.Logger = newConf.Logger
		return nil
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(conf *Config) error {
		conf.Logger = logger
		return nil
	}
}

func WithKey(key string) Option {
	return func(conf *Config) error {
		conf.Key = key
		return nil
	}
}

// TODO: database
// TODO: password
