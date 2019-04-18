package joe

import "github.com/stretchr/testify/mock"

// The inMemory type is fully tested in brain_test.go via TestBrain_Memory(…).

// memoryMock is used to test other components, especially when checking correct
// error handling.
type memoryMock struct {
	mock.Mock
}

func (m *memoryMock) Set(key, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *memoryMock) Get(key string) (string, bool, error) {
	args := m.Called(key)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *memoryMock) Delete(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

func (m *memoryMock) Memories() (mm map[string]string, err error) {
	args := m.Called()
	if x := args.Get(0); x != nil {
		mm = x.(map[string]string)
	}

	return mm, args.Error(1)
}

func (m *memoryMock) Close() error {
	args := m.Called()
	return args.Error(0)
}
