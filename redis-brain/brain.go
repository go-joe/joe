package redis

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/fgrosse/joe"
)

type Brain struct {
	logger *zap.Logger
	Client *redis.Client
	hkey   string
}

// TODO: is it actually a good idea to sprinkle the options pattern all over this?
func BrainOption(key string, opts ...Option) joe.Option {
	return func(b *joe.Bot) error {
		if b.Logger != nil {
			// TODO: actually this will overwrite a WithLogger option we got already
			opts = append(opts, WithLogger(b.Logger.Named("brain")))
		}

		brain, err := NewBrain(key, opts...)
		if err != nil {
			return err
		}

		b.Brain = brain
		return nil
	}
}

func NewBrain(addr string, opts ...Option) (*Brain, error) {
	brain := new(Brain)

	for _, opt := range opts {
		err := opt(brain)
		if err != nil {
			return nil, err
		}
	}

	if brain.logger == nil {
		brain.logger = zap.NewNop()
	}

	if brain.hkey == "" {
		brain.hkey = "joe-bot"
	}

	brain.logger.Debug("Connecting to redis memory",
		zap.String("addr", addr),
		zap.String("key", brain.hkey),
	)

	brain.Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := brain.Client.Ping().Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping redis")
	}

	brain.logger.Info("Memory initialized successfully")
	return brain, nil
}

func (b *Brain) Set(key, value string) error {
	b.logger.Debug("Writing data to memory", zap.String("key", key))
	resp := b.Client.HSet(b.hkey, key, value)
	return resp.Err()
}

func (b *Brain) Get(key string) (string, bool, error) {
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

func (b *Brain) Delete(key string) (bool, error) {
	b.logger.Debug("Deleting data from memory", zap.String("key", key))
	res, err := b.Client.HDel(b.hkey, key).Result()
	return res > 0, err
}

func (b *Brain) Memories() (map[string]string, error) {
	return b.Client.HGetAll(b.hkey).Result()
}

func (b *Brain) Close() error {
	return b.Client.Close()
}
