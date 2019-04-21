package joetest

import "testing"

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

func TestStorage_AssertEqualsError(t *testing.T) {
	mock := new(mockT)
	store := NewStorage(mock)
	store.AssertEquals("does-not-exist", "xxx")

	if len(mock.Errors) != 1 {
		t.Fatal("Expected one error but got none")
	}

	expected := `Expected storage to contain key "does-not-exist" but it does not`
	if mock.Errors[0] != expected {
		t.Errorf("Expected errors %q but got %q", expected, mock.Errors[0])
	}

	store.MustSet("test", "foo")
	store.AssertEquals("test", "bar")

	if len(mock.Errors) != 2 {
		t.Fatalf("Expected one error but got %d", len(mock.Errors))
	}

	expected = `Value of key "test" does not equal expected value
got:  "foo"
want: "bar"`
	if mock.Errors[1] != expected {
		t.Errorf("Expected errors %q but got %q", expected, mock.Errors[1])
	}
}
