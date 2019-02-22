package file

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/fgrosse/joe"
)

type Brain struct {
	path   string
	logger *zap.Logger

	mu   sync.RWMutex
	data map[string]string
}

func NewBrain(path string, opts ...Option) (*Brain, error) {
	brain := &Brain{
		path: path,
		data: map[string]string{},
	}

	for _, opt := range opts {
		err := opt(brain)
		if err != nil {
			return nil, err
		}
	}

	if brain.logger == nil {
		brain.logger = zap.NewNop()
	}

	brain.logger.Debug("Opening memory file", zap.String("path", path))
	f, err := os.Open(path)
	switch {
	case os.IsNotExist(err):
		brain.logger.Debug("File does not exist. Continuing with empty memory", zap.String("path", path))
	case err != nil:
		return nil, errors.Wrap(err, "failed to open file")
	default:
		brain.logger.Debug("Decoding JSON from memory file", zap.String("path", path))
		err := json.NewDecoder(f).Decode(&brain.data)
		_ = f.Close()
		if err != nil {
			return nil, errors.Wrap(err, "failed decode data as JSON")
		}
	}

	brain.logger.Info("Memory initialized successfully",
		zap.String("path", path),
		zap.Int("num_memories", len(brain.data)),
	)

	return brain, nil
}

func BrainOption(path string) joe.Option {
	return func(b *joe.Bot) error {
		var opts []Option
		if b.Logger != nil {
			opts = append(opts, WithLogger(b.Logger.Named("brain")))
		}

		brain, err := NewBrain(path, opts...)
		if err != nil {
			return err
		}

		b.Brain = brain
		return nil
	}
}

func (b *Brain) Set(key, value string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.data == nil {
		return errors.New("brain was already shut down")
	}

	b.logger.Debug("Writing data to memory", zap.String("key", key))
	b.data[key] = value
	err := b.persist()

	return err
}

func (b *Brain) Get(key string) (string, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.data == nil {
		return "", false, errors.New("brain was already shut down")
	}

	b.logger.Debug("Retrieving data from memory", zap.String("key", key))
	value, ok := b.data[key]
	return value, ok, nil
}

func (b *Brain) Delete(key string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.data == nil {
		return false, errors.New("brain was already shut down")
	}

	b.logger.Debug("Deleting data from memory", zap.String("key", key))
	_, ok := b.data[key]
	delete(b.data, key)
	err := b.persist()

	return ok, err
}

func (b *Brain) Memories() (map[string]string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.data == nil {
		return nil, errors.New("brain was already shut down")
	}

	m := make(map[string]string, len(b.data))
	for k, v := range b.data {
		m[k] = v
	}

	return m, nil
}

func (b *Brain) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.data == nil {
		return errors.New("brain was already closed")
	}

	b.logger.Debug("Shutting down brain")
	b.data = nil

	return nil
}

func (b *Brain) persist() error {
	f, err := os.OpenFile(b.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		return errors.Wrap(err, "failed to open file to persist data")
	}

	err = json.NewEncoder(f).Encode(b.data)
	if err != nil {
		_ = f.Close()
		return errors.Wrap(err, "failed to encode data as JSON")
	}

	err = f.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close file; data might not have been fully persisted to disk")
	}

	return nil
}
