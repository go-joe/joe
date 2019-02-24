package joe

import (
	"sync"

	"go.uber.org/zap"
)

type Brain struct {
	mu     sync.RWMutex
	memory Memory
	events EventEmitter
	logger *zap.Logger
}

type Memory interface {
	Set(key, value string) error
	Get(key string) (string, bool, error)
	Delete(key string) (bool, error)
	Memories() (map[string]string, error)
	Close() error
}

func NewBrain(m Memory, logger *zap.Logger, events EventEmitter) *Brain {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Brain{
		memory: m,
		logger: logger,
		events: events,
	}
}

func NewInMemoryBrain(logger *zap.Logger, events EventEmitter) *Brain {
	return NewBrain(newInMemory(), logger, events)
}

func newInMemory() *inMemory {
	return &inMemory{data: map[string]string{}}
}

func (b *Brain) Set(key, value string) error {
	b.mu.Lock()
	b.logger.Debug("Writing data to memory", zap.String("key", key))
	err := b.memory.Set(key, value)
	b.mu.Unlock()

	b.events.Emit(BrainMemoryEvent{Operation: "set", Key: key, Value: value})
	return err
}

func (b *Brain) Get(key string) (string, bool, error) {
	b.mu.RLock()
	b.logger.Debug("Retrieving data from memory", zap.String("key", key))
	value, ok, err := b.memory.Get(key)
	b.mu.RUnlock()

	b.events.Emit(BrainMemoryEvent{Operation: "get", Key: key, Value: value})
	return value, ok, err
}

func (b *Brain) Delete(key string) (bool, error) {
	b.mu.Lock()
	b.logger.Debug("Deleting data from memory", zap.String("key", key))
	ok, err := b.memory.Delete(key)
	b.mu.Unlock()

	b.events.Emit(BrainMemoryEvent{Operation: "del", Key: key})
	return ok, err
}

func (b *Brain) Memories() (map[string]string, error) {
	b.mu.RLock()
	data, err := b.memory.Memories()
	b.mu.RUnlock()

	return data, err
}

func (b *Brain) Close() error {
	b.mu.Lock()
	b.logger.Debug("Shutting down brain")
	err := b.memory.Close()
	b.mu.Unlock()

	return err
}

type inMemory struct {
	data map[string]string
}

func (m *inMemory) Set(key, value string) error {
	m.data[key] = value
	return nil
}

func (m *inMemory) Get(key string) (string, bool, error) {
	value, ok := m.data[key]
	return value, ok, nil
}

func (m *inMemory) Delete(key string) (bool, error) {
	_, ok := m.data[key]
	delete(m.data, key)
	return ok, nil
}

func (m *inMemory) Memories() (map[string]string, error) {
	data := make(map[string]string, len(m.data))
	for k, v := range m.data {
		data[k] = v
	}

	return data, nil
}

func (m *inMemory) Close() error {
	m.data = map[string]string{}
	return nil
}
