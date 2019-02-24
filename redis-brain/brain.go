package redis

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/fgrosse/joe"
)

type Config struct {
	Addr     string
	Key      string
	Password string
	DB       int
	Logger   *zap.Logger
}

type brain struct {
	logger *zap.Logger
	Client *redis.Client
	hkey   string
}

func Brain(addr string, opts ...Option) joe.Option {
	return func(b *joe.Bot) error {
		conf := Config{Addr: addr}
		for _, opt := range opts {
			err := opt(&conf)
			if err != nil {
				return err
			}
		}

		if conf.Logger == nil {
			conf.Logger = b.Logger
		}

		brain, err := NewBrain(conf)
		if err != nil {
			return err
		}

		b.Brain = brain
		return nil
	}
}

func NewBrain(conf Config) (joe.Brain, error) {
	if conf.Logger == nil {
		conf.Logger = zap.NewNop()
	}

	if conf.Key == "" {
		conf.Key = "joe-bot"
	}

	brain := &brain{
		logger: conf.Logger,
		hkey:   conf.Key,
	}

	brain.logger.Debug("Connecting to redis memory",
		zap.String("addr", conf.Addr),
		zap.String("key", brain.hkey),
	)

	brain.Client = redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})

	_, err := brain.Client.Ping().Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping redis")
	}

	brain.logger.Info("Memory initialized successfully")
	return brain, nil
}

func (b *brain) Set(key, value string) error {
	b.logger.Debug("Writing data to memory", zap.String("key", key))
	resp := b.Client.HSet(b.hkey, key, value)
	return resp.Err()
}

func (b *brain) Get(key string) (string, bool, error) {
	b.logger.Debug("Retrieving data from memory", zap.String("key", key))
	res, err := b.Client.HGet(b.hkey, key).Result()
	switch {
	case err == redis.Nil:
		return "", false, nil
	case err != nil:
		return "", false, err
	default:
		return res, true, nil
	}
}

func (b *brain) Delete(key string) (bool, error) {
	b.logger.Debug("Deleting data from memory", zap.String("key", key))
	res, err := b.Client.HDel(b.hkey, key).Result()
	return res > 0, err
}

func (b *brain) Memories() (map[string]string, error) {
	return b.Client.HGetAll(b.hkey).Result()
}

func (b *brain) Close() error {
	return b.Client.Close()
}
