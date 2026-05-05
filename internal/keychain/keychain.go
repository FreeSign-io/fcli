// Package keychain stores and retrieves FreeSign API tokens from the OS-native
// secret store using zalando/go-keyring, which is pure Go and shells out to:
//
//   - macOS:   /usr/bin/security
//   - Linux:   org.freedesktop.secrets via dbus
//   - Windows: wincred
//
// We use it instead of 99designs/keyring so the CGO_ENABLED=0 cross-compiled
// binaries still talk to the real OS keychain on the user's machine. Tokens
// are keyed by profile name. We never persist tokens in the on-disk config
// file.
package keychain

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const serviceName = "fcli"

// ErrNotFound is returned when no token exists for the given profile.
var ErrNotFound = errors.New("token not found")

// Set stores the token for a profile, replacing any existing value.
func Set(profile, token string) error {
	if err := keyring.Set(serviceName, profile, token); err != nil {
		return fmt.Errorf("write keychain: %w", err)
	}
	return nil
}

// Get returns the token for a profile, or ErrNotFound.
func Get(profile string) (string, error) {
	v, err := keyring.Get(serviceName, profile)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("read keychain: %w", err)
	}
	return v, nil
}

// Delete removes a profile's token. Missing tokens are a no-op.
func Delete(profile string) error {
	if err := keyring.Delete(serviceName, profile); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("remove keychain entry: %w", err)
	}
	return nil
}
