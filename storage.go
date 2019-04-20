package joe

import (
	"encoding/json"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Storage struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	memory  Memory
	encoder MemoryEncoder
}

// The Memory interface allows the bot to persist data as key-value pairs.
// The default implementation of the Memory is to store all keys and values in
// a map (i.e. in-memory). Other implementations typically offer actual long term
// persistence into a file or to redis.
type Memory interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, bool, error)
	Delete(key string) (bool, error)
	Keys() ([]string, error)
	Close() error
}

type MemoryEncoder interface {
	Encode(value interface{}) ([]byte, error)
	Decode(data []byte, target interface{}) error
}

type inMemory struct {
	data map[string][]byte
}

type jsonEncoder struct{}

func NewStorage(logger *zap.Logger) *Storage {
	return &Storage{
		logger:  logger,
		memory:  newInMemory(),
		encoder: new(jsonEncoder),
	}
}

// SetMemory assigns a different Memory implementation.
func (s *Storage) SetMemory(m Memory) {
	s.mu.Lock()
	s.memory = m
	s.mu.Unlock()
}

// SetMemoryEncoder assigns a different MemoryEncoder.
func (s *Storage) SetMemoryEncoder(enc MemoryEncoder) {
	s.mu.Lock()
	s.encoder = enc
	s.mu.Unlock()
}

// Keys returns all keys known to the Memory.
func (s *Storage) Keys() ([]string, error) {
	s.mu.RLock()
	keys, err := s.memory.Keys()
	s.mu.RUnlock()

	sort.Strings(keys)
	return keys, err
}

func (s *Storage) Set(key string, value interface{}) error {
	data, err := s.encoder.Encode(value)
	if err != nil {
		return errors.Wrap(err, "encode data")
	}

	s.mu.Lock()
	s.logger.Debug("Writing data to memory", zap.String("key", key))
	err = s.memory.Set(key, data)
	s.mu.Unlock()

	return err
}

func (s *Storage) Get(key string, value interface{}) (bool, error) {
	s.mu.RLock()
	s.logger.Debug("Retrieving data from memory", zap.String("key", key))
	data, ok, err := s.memory.Get(key)
	s.mu.RUnlock()
	if err != nil {
		return false, errors.WithStack(err)
	}

	if !ok || value == nil {
		return ok, nil
	}

	err = s.encoder.Decode(data, value)
	if err != nil {
		return false, errors.Wrap(err, "decode data")
	}

	return true, nil
}

func (s *Storage) Delete(key string) (bool, error) {
	s.mu.Lock()
	s.logger.Debug("Deleting data from memory", zap.String("key", key))
	ok, err := s.memory.Delete(key)
	s.mu.Unlock()

	return ok, err
}

func (s *Storage) Close() error {
	s.mu.Lock()
	err := s.memory.Close()
	s.mu.Unlock()
	return err
}

func newInMemory() *inMemory {
	return &inMemory{data: map[string][]byte{}}
}

func (m *inMemory) Set(key string, value []byte) error {
	m.data[key] = value
	return nil
}

func (m *inMemory) Get(key string) ([]byte, bool, error) {
	value, ok := m.data[key]
	return value, ok, nil
}

func (m *inMemory) Delete(key string) (bool, error) {
	_, ok := m.data[key]
	delete(m.data, key)
	return ok, nil
}

func (m *inMemory) Keys() ([]string, error) {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}

	return keys, nil
}

func (m *inMemory) Close() error {
	m.data = map[string][]byte{}
	return nil
}

func (jsonEncoder) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (jsonEncoder) Decode(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}
