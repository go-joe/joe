package joetest

import "fmt"

type mockT struct {
	Errors []string
	failed bool
	fatal  bool
}

func (m *mockT) Logf(string, ...interface{}) {}

func (m *mockT) Errorf(msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	m.Errors = append(m.Errors, msg)
}

func (m *mockT) Fail() {
	m.failed = true
}

func (m *mockT) Failed() bool {
	return m.failed
}

func (m *mockT) Fatal(args ...interface{}) {
	m.failed = true
	m.fatal = true
}

func (*mockT) Name() string {
	return "mock"
}

func (m *mockT) FailNow() {
	m.Fatal()
}

func (*mockT) Helper() {}
