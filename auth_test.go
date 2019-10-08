package joe_test

import (
	"errors"
	"testing"

	"github.com/go-joe/joe"
	"github.com/go-joe/joe/joetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)
	userID := "fgrosse"

	// Initially the user should have no permissions whatsoever
	err := auth.CheckPermission("test.foo", userID)
	require.Equal(t, joe.ErrNotAllowed, err)

	// Granting the empty scope is likely an error and thus should result in an error
	_, err = auth.Grant("", userID)
	require.EqualError(t, err, "scope cannot be empty")
	err = auth.CheckPermission("", userID)
	require.Equal(t, joe.ErrNotAllowed, err)

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
	require.Equal(t, joe.ErrNotAllowed, err)

	// foo is also not contained in the test.foo scope so this should be denied.
	err = auth.CheckPermission("foo", userID)
	require.Equal(t, joe.ErrNotAllowed, err)

	// Even though test.foo and test.bar share a common prefix this scope is not entirely
	// contained in the granted scope so this should be denied.
	err = auth.CheckPermission("test.bar", userID)
	require.Equal(t, joe.ErrNotAllowed, err)
}

func TestAuth_GrantIsIdempotent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	// Lets assume dave already has permissions ot open the pod bay doors we want
	// to make sure we will not append the same permissions multiple times.
	existingPermissions := []string{"open_pod_bay_doors", "foo.bar"}
	store.MustSet("joe.permissions.dave", existingPermissions)

	ok, err := auth.Grant("open_pod_bay_doors", "dave")
	require.NoError(t, err)
	assert.False(t, ok)

	store.AssertEquals("joe.permissions.dave", existingPermissions)
}

func TestAuth_GrantWiderScope(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	store.MustSet("joe.permissions.fgrosse", []string{"foo.bar.baz", "test"})

	ok, err := auth.Grant("foo", "fgrosse")
	require.NoError(t, err)
	assert.True(t, ok)

	store.AssertEquals("joe.permissions.fgrosse", []string{"test", "foo"})
}

func TestAuth_GrantSmallerScope(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	store.MustSet("joe.permissions.fgrosse", []string{"foo", "test"})

	ok, err := auth.Grant("foo.bar.baz", "fgrosse")
	require.NoError(t, err)
	assert.False(t, ok)

	store.AssertEquals("joe.permissions.fgrosse", []string{"foo", "test"})
}

func TestAuth_Revoke(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	store.MustSet("joe.permissions.fgrosse", []string{"foo.bar", "test"})

	ok, err := auth.Revoke("foo.bar", "fgrosse")
	assert.NoError(t, err)
	assert.True(t, ok)

	store.AssertEquals("joe.permissions.fgrosse", []string{"test"})
}

func TestAuth_RevokeNonExistingScope(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	store.MustSet("joe.permissions.fgrosse", []string{"test"})

	ok, err := auth.Revoke("foo.bar", "fgrosse")
	assert.NoError(t, err)
	assert.False(t, ok)

	store.AssertEquals("joe.permissions.fgrosse", []string{"test"})
}

func TestAuth_RevokeWiderScope(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	store.MustSet("joe.permissions.fgrosse", []string{"foo"})

	ok, err := auth.Revoke("foo.bar", "fgrosse")
	assert.EqualError(t, err, `cannot revoke scope "foo.bar" because the user still has the more general scope "foo"`)
	assert.False(t, ok)
}

func TestAuth_RevokeEmptyScope(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	ok, err := auth.Revoke("", "fgrosse")
	assert.EqualError(t, err, "scope cannot be empty")
	assert.False(t, ok)
}

func TestAuth_RevokeLastScope(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	store.MustSet("joe.permissions.fgrosse", []string{"test"})

	ok, err := auth.Revoke("test", "fgrosse")
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = store.Get("joe.permissions.fgrosse", nil)
	require.NoError(t, err)
	assert.False(t, ok, "storage should no longer contain any permissions for this user")
}

func TestAuth_RevokeNoOldScopes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	ok, err := auth.Revoke("test", "fgrosse")
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, err = store.Get("joe.permissions.fgrosse", nil)
	require.NoError(t, err)
	assert.False(t, ok, "storage still not contain any permissions for this user")
}

func TestAuth_CheckPermission_Errors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	store := joetest.NewStorage(t)
	store.SetMemory(mem)
	auth := joe.NewAuth(logger, store.Storage)

	mem.On("Get", "joe.permissions.test").Return(nil, false, errors.New("that didn't work"))
	err := auth.CheckPermission("xxx", "test")
	assert.EqualError(t, err, "failed to load user permissions: that didn't work")

	mem = new(memoryMock)
	store.SetMemory(mem)

	mem.On("Get", "joe.permissions.test").Return([]byte("nope!"), true, nil)
	err = auth.CheckPermission("xxx", "test")
	assert.EqualError(t, err, "failed to load user permissions: decode data: invalid character 'o' in literal null (expecting 'u')")
}

func TestAuth_Grant_Errors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	store := joetest.NewStorage(t)
	store.SetMemory(mem)
	auth := joe.NewAuth(logger, store.Storage)

	mem.On("Get", "joe.permissions.test").Return(nil, false, errors.New("that didn't work"))
	_, err := auth.Grant("xxx", "test")
	assert.EqualError(t, err, "failed to load user permissions: that didn't work")

	mem = new(memoryMock)
	store.SetMemory(mem)

	mem.On("Get", "joe.permissions.test").Return(nil, false, nil)
	mem.On("Set", "joe.permissions.test", []byte(`["xxx"]`)).Return(errors.New("not today"))
	_, err = auth.Grant("xxx", "test")
	assert.EqualError(t, err, "failed to update user permissions: not today")
}

func TestAuth_Revoke_Errors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mem := new(memoryMock)
	store := joetest.NewStorage(t)
	store.SetMemory(mem)
	auth := joe.NewAuth(logger, store.Storage)

	mem.On("Get", "joe.permissions.test").Return(nil, false, errors.New("that didn't work"))
	_, err := auth.Revoke("xxx", "test")
	assert.EqualError(t, err, "failed to load user permissions: that didn't work")

	mem = new(memoryMock)
	store.SetMemory(mem)

	mem.On("Get", "joe.permissions.test").Return([]byte(`["foo", "bar"]`), true, nil)
	mem.On("Set", "joe.permissions.test", []byte(`["bar"]`)).Return(errors.New("not today"))
	_, err = auth.Revoke("foo", "test")
	assert.EqualError(t, err, "failed to update user permissions: not today")

	mem = new(memoryMock)
	store.SetMemory(mem)

	mem.On("Get", "joe.permissions.test").Return([]byte(`["foo"]`), true, nil)
	mem.On("Delete", "joe.permissions.test").Return(false, errors.New("not today"))
	_, err = auth.Revoke("foo", "test")
	assert.EqualError(t, err, "failed to delete last user permission: not today")
}

func TestAuth_GetUsers(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	mustHaveUsers := map[string][]string{
		"dave": {"bot.scopeA", "bot.scopeB"},
		"john": {"bot.scopeC", "bot.scopeD"},
	}
	for user, perms := range mustHaveUsers {
		for _, scope := range perms {
			auth.Grant(scope, user)
		}
	}

	// GetUsers() should return a list of userIDs
	users, err := auth.Users()
	require.NoError(t, err)
	for user := range mustHaveUsers {
		require.Contains(t, users, user)
	}
}

func TestAuth_GetUserPermissions(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := joetest.NewStorage(t)
	auth := joe.NewAuth(logger, store.Storage)

	mustHavePermissions := map[string][]string{
		"dave": {"bot.scopeA", "bot.scopeB"},
		"john": {"bot.scopeC", "bot.scopeD"},
	}
	for user, perms := range mustHavePermissions {
		for _, scope := range perms {
			auth.Grant(scope, user)
		}
	}

	// GetUserPermissions() should return all permission scopes for a user
	for _, user := range []string{"dave", "john"} {
		permissions, err := auth.UserPermissions(user)
		require.NoError(t, err)
		for _, scope := range permissions {
			err = auth.CheckPermission(scope, user)
			require.NoError(t, err)
		}
	}

}

type memoryMock struct {
	mock.Mock
}

func (m *memoryMock) Set(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *memoryMock) Get(key string) (data []byte, ok bool, err error) {
	args := m.Called(key)
	if x := args.Get(0); x != nil {
		data = x.([]byte)
	}

	return data, args.Bool(1), args.Error(2)
}

func (m *memoryMock) Delete(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

func (m *memoryMock) Keys() (keys []string, err error) {
	args := m.Called()
	if x := args.Get(0); x != nil {
		keys = x.([]string)
	}

	return keys, args.Error(1)
}

func (m *memoryMock) Close() error {
	args := m.Called()
	return args.Error(0)
}
