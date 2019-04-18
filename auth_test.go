package joe

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := newInMemory()
	auth := NewAuth(logger, mem)
	userID := "fgrosse"

	// Initially the user should have no permissions whatsoever
	err := auth.CheckPermission("test.foo", userID)
	require.Equal(t, ErrNotAllowed, err)

	// Granting the empty scope is likely an error and thus should result in an error
	err = auth.Grant("", userID)
	require.EqualError(t, err, "scope cannot be empty")
	err = auth.CheckPermission("", userID)
	require.Equal(t, ErrNotAllowed, err)

	// Grant the test.foo scope
	err = auth.Grant("test.foo", userID)
	require.NoError(t, err)

	// The user has exactly the test.foo scope and should be granted access.
	err = auth.CheckPermission("test.foo", userID)
	require.NoError(t, err)

	// test.foo.bar is contained in the test.foo scope and the user should be granted access.
	err = auth.CheckPermission("test.foo.bar", userID)
	require.NoError(t, err)

	// test is not contained in the test.foo scope so this should be denied.
	err = auth.CheckPermission("test", userID)
	require.Equal(t, ErrNotAllowed, err)

	// foo is also not contained in the test.foo scope so this should be denied.
	err = auth.CheckPermission("foo", userID)
	require.Equal(t, ErrNotAllowed, err)

	// Even though test.foo and test.bar share a common prefix this scope is not entirely
	// contained in the granted scope so this should be denied.
	err = auth.CheckPermission("test.bar", userID)
	require.Equal(t, ErrNotAllowed, err)
}

func TestAuth_CheckPermission_Errors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	m := new(memoryMock)
	auth := NewAuth(logger, m)

	m.On("Get", "joe.permissions.test").Return("", false, errors.New("that didn't work"))
	err := auth.CheckPermission("xxx", "test")
	assert.EqualError(t, err, "failed to load user permissions: that didn't work")

	m = new(memoryMock)
	auth = NewAuth(logger, m)

	m.On("Get", "joe.permissions.test").Return("nope!", true, nil)
	err = auth.CheckPermission("xxx", "test")
	assert.EqualError(t, err, "failed to decode user permissions as JSON: invalid character 'o' in literal null (expecting 'u')")
}

func TestAuth_Grant_Errors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	m := new(memoryMock)
	auth := NewAuth(logger, m)

	m.On("Get", "joe.permissions.test").Return("", false, errors.New("that didn't work"))
	err := auth.Grant("xxx", "test")
	assert.EqualError(t, err, "failed to load user permissions: that didn't work")

	m = new(memoryMock)
	auth = NewAuth(logger, m)

	m.On("Get", "joe.permissions.test").Return("", false, nil)
	m.On("Set", "joe.permissions.test", `["xxx"]`).Return(errors.New("not today"))
	err = auth.Grant("xxx", "test")
	assert.EqualError(t, err, "failed to store user permissions: not today")
}
