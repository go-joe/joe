package joe

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ErrNotAllowed is returned if the user is not allowed access to a specific scope.
const ErrNotAllowed = Error("not allowed")

type Auth struct {
	logger *zap.Logger
	memory Memory
}

func NewAuth(logger *zap.Logger, memory Memory) *Auth {
	return &Auth{
		logger: logger,
		memory: memory,
	}
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

// Grant adds a new permission scope to the given user. When a scope was granted
// to a specific user it can be checked later via CheckPermission(…). The empty
// scope cannot be granted and trying to do so will result in an error. If you
// want to grant access to all scopes you should prefix them with a common scope
// such as "root." or "api.".
func (a *Auth) Grant(scope, userID string) error {
	if scope == "" {
		return errors.New("scope cannot be empty")
	}

	key := a.permissionsKey(userID)
	permissions, err := a.loadPermissions(key)
	if err != nil {
		return errors.WithStack(err)
	}

	permissions = append(permissions, scope)
	data, err := json.Marshal(permissions)
	if err != nil {
		return errors.Wrap(err, "failed to encode permissions as JSON")
	}

	a.logger.Info("Granting user permission",
		zap.String("scope", scope),
		zap.String("userID", userID),
	)

	err = a.memory.Set(key, string(data))
	if err != nil {
		return errors.Wrap(err, "failed to store user permissions")
	}

	return nil
}

func (a *Auth) permissionsKey(userID string) string {
	return "joe.permissions." + userID
}
