package joetest

import (
	"reflect"

	"github.com/go-joe/joe"
	"go.uber.org/zap/zaptest"
)

// Storage wraps a joe.Storage for unit testing purposes.
type Storage struct {
	*joe.Storage
	T TestingT
}

// NewStorage creates a new Storage.
func NewStorage(t TestingT) *Storage {
	logger := zaptest.NewLogger(t)
	return &Storage{
		Storage: joe.NewStorage(logger),
		T:       t,
	}
}

// MustSet assigns the value to the given key and fails the test immediately if
// there was an error.
func (s *Storage) MustSet(key string, value interface{}) {
	err := s.Set(key, value)
	if err != nil {
		s.T.Fatal("Failed to set key in storage:", err)
	}
}

// AssertEquals checks that the actual value under the given key equals an
// expected value.
func (s *Storage) AssertEquals(key string, expectedVal interface{}) {
	typ := reflect.TypeOf(expectedVal)
	actual := reflect.New(typ)
	ok, err := s.Get(key, actual.Interface())
	if err != nil {
		s.T.Errorf("Error while getting key %q from storage: %v", key, err)
		return
	}

	if !ok {
		s.T.Errorf("Expected storage to contain key %q but it does not", key)
		return
	}

	actualVal := actual.Elem().Interface()
	if !reflect.DeepEqual(expectedVal, actualVal) {
		s.T.Errorf("Value of key %q does not equal expected value\ngot:  %#v\nwant: %#v", key, actualVal, expectedVal)
	}
}
