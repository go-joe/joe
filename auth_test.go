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
	_, err = auth.Grant("", userID)
	require.EqualError(t, err, "scope cannot be empty")
	err = auth.CheckPermission("", userID)
	require.Equal(t, ErrNotAllowed, err)

	// Grant the test.foo scope
	ok, err := auth.Grant("test.foo", userID)
	require.NoError(t, err)
	assert.True(t, ok)

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

func TestAuth_GrantIsIdempotent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	auth := NewAuth(logger, mem)

	// Lets assume day already has permissions ot open the pod bay doors we want
	// to make sure we will not append the same permissions multiple times.
	mem.On("Get", "joe.permissions.dave").Return(`["open_pod_bay_doors","foo.bar"]`, true, nil)

	// We should not see a call to mem.Set(…) since no permissions need to be updated

	ok, err := auth.Grant("open_pod_bay_doors", "dave")
	require.NoError(t, err)
	assert.False(t, ok)

	mem.AssertExpectations(t)
}

func TestAuth_GrantWiderScope(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	auth := NewAuth(logger, mem)

	// Lets assume day already has very specific permissions and now we are adding
	// a wider scope that contains the original permissions.
	mem.On("Get", "joe.permissions.fgrosse").Return(`["foo.bar.baz", "test"]`, true, nil)
	mem.On("Set", "joe.permissions.fgrosse", `["test","foo"]`).Return(nil)

	ok, err := auth.Grant("foo", "fgrosse")
	require.NoError(t, err)
	assert.True(t, ok)

	mem.AssertExpectations(t)
}

func TestAuth_GrantSmallerScope(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	auth := NewAuth(logger, mem)

	mem.On("Get", "joe.permissions.fgrosse").Return(`["foo", "test"]`, true, nil)

	// We should not see a call to mem.Set(…) since no permissions were actually added

	ok, err := auth.Grant("foo.bar.baz", "fgrosse")
	require.NoError(t, err)
	assert.False(t, ok)

	mem.AssertExpectations(t)
}

func TestAuth_Revoke(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	auth := NewAuth(logger, mem)

	mem.On("Get", "joe.permissions.fgrosse").Return(`["test"]`, true, nil)
	ok, err := auth.Revoke("foo.bar", "fgrosse")
	assert.NoError(t, err)
	assert.False(t, ok)
	mem.AssertExpectations(t)

	mem = new(memoryMock)
	auth = NewAuth(logger, mem)

	mem.On("Get", "joe.permissions.fgrosse").Return(`["foo.bar", "test"]`, true, nil)
	mem.On("Set", "joe.permissions.fgrosse", `["test"]`).Return(nil)

	ok, err = auth.Revoke("foo.bar", "fgrosse")
	assert.NoError(t, err)
	assert.True(t, ok)

	// TODO: what if we revoke "foo.bar" and the user has the wider "foo" scope?

	mem.AssertExpectations(t)
}

func TestAuth_RevokeWiderScope(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	auth := NewAuth(logger, mem)

	mem.On("Get", "joe.permissions.fgrosse").Return(`["foo"]`, true, nil)

	ok, err := auth.Revoke("foo.bar", "fgrosse")
	assert.EqualError(t, err, `cannot revoke scope "foo.bar" because the user still has the more general scope "foo"`)
	assert.False(t, ok)
}

func TestAuth_RevokeEmptyScope(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	auth := NewAuth(logger, mem)

	ok, err := auth.Revoke("", "fgrosse")
	assert.EqualError(t, err, "scope cannot be empty")
	assert.False(t, ok)
}

func TestAuth_CheckPermission_Errors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	auth := NewAuth(logger, mem)

	mem.On("Get", "joe.permissions.test").Return("", false, errors.New("that didn't work"))
	err := auth.CheckPermission("xxx", "test")
	assert.EqualError(t, err, "failed to load user permissions: that didn't work")

	mem = new(memoryMock)
	auth = NewAuth(logger, mem)

	mem.On("Get", "joe.permissions.test").Return("nope!", true, nil)
	err = auth.CheckPermission("xxx", "test")
	assert.EqualError(t, err, "failed to decode user permissions as JSON: invalid character 'o' in literal null (expecting 'u')")
}

func TestAuth_Grant_Errors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	auth := NewAuth(logger, mem)

	mem.On("Get", "joe.permissions.test").Return("", false, errors.New("that didn't work"))
	_, err := auth.Grant("xxx", "test")
	assert.EqualError(t, err, "failed to load user permissions: that didn't work")

	mem = new(memoryMock)
	auth = NewAuth(logger, mem)

	mem.On("Get", "joe.permissions.test").Return("", false, nil)
	mem.On("Set", "joe.permissions.test", `["xxx"]`).Return(errors.New("not today"))
	_, err = auth.Grant("xxx", "test")
	assert.EqualError(t, err, "failed to store user permissions: not today")
}
