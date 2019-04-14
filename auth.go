package joe

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// ErrNotAllowed is returned if the user is not allowed access to a specific scope.
const ErrNotAllowed = Error("not allowed")

type auth struct {
	mu          sync.RWMutex
	permissions map[string][]string // maps user IDs to a list of granted scopes
}

func newAuth() *auth {
	return &auth{permissions: map[string][]string{}}
}

// CheckPermissions checks if a user has permissions to access a resource under
// a given scope. If the user is not permitted access this function returns
// ErrNotAllowed.
//
// Scopes are interpreted in a hierarchical way where scope A can be contained
// in scope B if B is a prefix to A. For example, you can check if a user is
// allowed to read or write from the "Example" API by checking the
// "api.example.read" or "api.example.write" scope. When you grant the scope to
// a user you can now either decide only to grant the very specific
// "api.example.read" scope which means the user will not have write permissions
// or you can allow people write-only access via "api.example.write".
// Alternatively you can also grant any access to the Example API via "api.example"
// which includes both the read and write scope beneath it. If you choose to you
// could also allow even more general access to everything in the api via the
// "api" scope. The empty scope "" cannot be granted and will thus always return
// an error in the permission check.
func (a *auth) CheckPermission(scope, userID string) error {
	a.mu.RLock()
	permissions := a.permissions[userID]
	a.mu.RUnlock()

	for _, p := range permissions {
		if strings.HasPrefix(scope, p) {
			return nil
		}
	}

	return ErrNotAllowed
}

// Grant adds a new permission scope to the given user. When a scope was granted
// to a specific user it can be checked later via CheckPermission(…). The empty
// scope cannot be granted and trying to do so will result in an error. If you
// want to grant access to all scopes you should prefix them with a common scope
// such as "root." or "api.".
func (a *auth) Grant(scope, userID string) error {
	if scope == "" {
		return errors.New("scope cannot be empty")
	}

	a.mu.Lock()
	a.permissions[userID] = append(a.permissions[userID], scope)
	a.mu.Unlock()
	return nil
}
