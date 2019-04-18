package joe

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ErrNotAllowed is returned if the user is not allowed access to a specific scope.
const ErrNotAllowed = Error("not allowed")

// Auth implements logic to add user authorization checks to your bot.
type Auth struct {
	logger *zap.Logger
	memory Memory
}

// NewAuth creates a new Auth instance.
func NewAuth(logger *zap.Logger, memory Memory) *Auth {
	return &Auth{
		logger: logger,
		memory: memory,
	}
}

// CheckPermission checks if a user has permissions to access a resource under a
// given scope. If the user is not permitted access this function returns
// ErrNotAllowed.
//
// Scopes are interpreted in a hierarchical way where scope A can contain scope B
// if A is a prefix to B. For example, you can check if a user is allowed to
// read or write from the "Example" API by checking the "api.example.read" or
// "api.example.write" scope. When you grant the scope to a user you can now
// either decide only to grant the very specific "api.example.read" scope which
// means the user will not have write permissions or you can allow people
// write-only access via "api.example.write".
//
// Alternatively you can also grant any access to the Example API via "api.example"
// which includes both the read and write scope beneath it. If you choose to you
// could also allow even more general access to everything in the api via the
// "api" scope. The empty scope "" cannot be granted and will thus always return
// an error in the permission check.
func (a *Auth) CheckPermission(scope, userID string) error {
	key := a.permissionsKey(userID)
	permissions, err := a.loadPermissions(key)
	if err != nil {
		return errors.WithStack(err)
	}

	a.logger.Debug("Checking user permissions",
		zap.String("requested_scope", scope),
		zap.String("user_id", userID),
	)

	for _, p := range permissions {
		if strings.HasPrefix(scope, p) {
			return nil
		}
	}

	return ErrNotAllowed
}

func (a *Auth) loadPermissions(key string) ([]string, error) {
	data, ok, err := a.memory.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load user permissions")
	}

	if !ok {
		return nil, nil
	}

	var permissions []string
	err = json.NewDecoder(strings.NewReader(data)).Decode(&permissions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode user permissions as JSON")
	}

	return permissions, nil
}

// Grant adds a permission scope to the given user. When a scope was granted
// to a specific user it can be checked later via CheckPermission(…).
// The returned boolean indicates whether the scope was actually added (i.e. true)
// or the user already had the granted scope (false).
//
// Note that granting a scope is an idempotent operations so granting the same
// scope multiple times is a safe operation and will not change the internal
// permissions that are written to the Memory.
//
// The empty scope cannot be granted and trying to do so will result in an error.
// If you want to grant access to all scopes you should prefix them with a
// common scope such as "root." or "api.".
func (a *Auth) Grant(scope, userID string) (bool, error) {
	if scope == "" {
		return false, errors.New("scope cannot be empty")
	}

	key := a.permissionsKey(userID)
	oldPermissions, err := a.loadPermissions(key)
	if err != nil {
		return false, errors.WithStack(err)
	}

	newPermissions := make([]string, 0, len(oldPermissions)+1)
	for _, p := range oldPermissions {
		if strings.HasPrefix(scope, p) {
			// The user already has this or a scope that "contains" it
			return false, nil
		}

		if !strings.HasPrefix(p, scope) {
			newPermissions = append(newPermissions, p)
		}
	}

	newPermissions = append(newPermissions, scope)
	data, err := json.Marshal(newPermissions)
	if err != nil {
		return false, errors.Wrap(err, "failed to encode permissions as JSON")
	}

	a.logger.Info("Granting user permission",
		zap.String("userID", userID),
		zap.String("scope", scope),
	)

	err = a.memory.Set(key, string(data))
	if err != nil {
		return false, errors.Wrap(err, "failed to update user permissions")
	}

	return true, nil
}

// Revoke removes a previously grantde permission from a user. If the user does
// not currently have the revoked scope this function returns false and no error.
//
// If you are trying to revoke a permission but the user was previously granted
// a scope that contains the revoked scope this function returns an error.
func (a *Auth) Revoke(scope, userID string) (bool, error) {
	if scope == "" {
		return false, errors.New("scope cannot be empty")
	}

	key := a.permissionsKey(userID)
	oldPermissions, err := a.loadPermissions(key)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if len(oldPermissions) == 0 {
		return false, nil
	}

	var revoked bool
	newPermissions := make([]string, 0, len(oldPermissions))
	for _, p := range oldPermissions {
		if p == scope {
			revoked = true
			continue
		}

		if strings.HasPrefix(scope, p) {
			return false, errors.Errorf("cannot revoke scope %q because the user still has the more general scope %q", scope, p)
		}

		newPermissions = append(newPermissions, p)
	}

	if !revoked {
		return false, nil
	}

	a.logger.Info("Revoking user permission",
		zap.String("userID", userID),
		zap.String("scope", scope),
	)

	if len(newPermissions) == 0 {
		_, err := a.memory.Delete(key)
		if err != nil {
			return false, errors.Wrap(err, "failed to delete last user permission")
		}

		return true, nil
	}

	data, err := json.Marshal(newPermissions)
	if err != nil {
		return false, errors.Wrap(err, "failed to encode permissions as JSON")
	}

	err = a.memory.Set(key, string(data))
	if err != nil {
		return false, errors.Wrap(err, "failed to update user permissions")
	}

	return true, nil
}

func (a *Auth) permissionsKey(userID string) string {
	return "joe.permissions." + userID
}
