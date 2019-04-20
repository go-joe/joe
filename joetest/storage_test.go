package joetest

import (
	"testing"
)

func TestStorage(t *testing.T) {
	type CustomType struct{ N int }

	cases := []struct {
		key   string
		value interface{}
	}{
		{"test.string", "foobar"},
		{"test.bool", true},
		{"test.int", 42},
		{"test.float", 3.14159265359},
		{"test.string_slice", []string{"foo", "bar"}},
		{"test.struct", CustomType{1234}},
		{"test.ptr", &CustomType{1234}},
	}

	for _, c := range cases {
		store := NewStorage(t)
		store.MustSet(c.key, c.value)
		store.AssertEquals(c.key, c.value)
	}
}
