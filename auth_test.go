package joe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	auth := newAuth()
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
