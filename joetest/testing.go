// Package joetest implements helpers to implement unit tests for bots.
package joetest

// TestingT is the minimum required subset of the testing API used in the
// joetest package. TestingT is implemented both by *testing.T and *testing.B.
type TestingT interface {
	Logf(string, ...interface{})
	Errorf(string, ...interface{})
	Fail()
	Failed() bool
	Name() string
	FailNow()
}
