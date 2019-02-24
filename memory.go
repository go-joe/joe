package joe

type Memory interface {
	Set(key, value string) error
	Get(key string) (string, bool, error)
	Delete(key string) (bool, error)
	Memories() (map[string]string, error)
	Close() error
}

type inMemory struct {
	data map[string]string
}

func newInMemory() *inMemory {
	return &inMemory{data: map[string]string{}}
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
