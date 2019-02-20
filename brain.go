package joe

import "sync"

type Brain interface {
	Set(key, value string) error
	Get(key string) (string, bool, error)
	Delete(key string) (bool, error)
	Memories() (map[string]string, error)
	Close() error
}

type InMemoryBrain struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewInMemoryBrain() *InMemoryBrain {
	return &InMemoryBrain{
		data: map[string]string{},
	}
}

func (b *InMemoryBrain) Set(key, value string) error {
	b.mu.Lock()
	b.data[key] = value
	b.mu.Unlock()

	return nil
}

func (b *InMemoryBrain) Get(key string) (string, bool, error) {
	b.mu.RLock()
	value, ok := b.data[key]
	b.mu.RUnlock()

	return value, ok, nil
}

func (b *InMemoryBrain) Delete(key string) (bool, error) {
	b.mu.Lock()
	_, ok := b.data[key]
	delete(b.data, key)
	b.mu.Unlock()

	return ok, nil
}

func (b *InMemoryBrain) Close() error {
	b.mu.Lock()
	b.data = map[string]string{}
	b.mu.Unlock()

	return nil
}

func (b *InMemoryBrain) Memories() (map[string]string, error) {
	b.mu.RLock()
	m := make(map[string]string, len(b.data))
	for k, v := range b.data {
		m[k] = v
	}
	b.mu.RUnlock()

	return m, nil
}
