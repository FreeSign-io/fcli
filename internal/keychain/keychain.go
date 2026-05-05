// Package keychain stores and retrieves FreeSign API tokens from the OS-native
// secret store (macOS Keychain, Linux Secret Service, Windows Credential Manager)
// via 99designs/keyring.
//
// Tokens are keyed by profile name. We never persist tokens in the on-disk
// config file.
package keychain

import (
	"errors"
	"fmt"

	"github.com/99designs/keyring"
)

const serviceName = "fcli"

// ErrNotFound is returned when no token exists for the given profile.
var ErrNotFound = errors.New("token not found")

func open() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: serviceName,

		// macOS: use the user's login keychain.
		KeychainName:                   "login",
		KeychainTrustApplication:       true,
		KeychainAccessibleWhenUnlocked: true,

		// Linux: prefer Secret Service (gnome-keyring / kwallet); fall back to
		// the encrypted file backend so headless servers still work.
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.SecretServiceBackend,
			keyring.WinCredBackend,
			keyring.FileBackend,
		},

		// File backend (last-resort): require a passphrase via env var so
		// scripted use stays non-interactive.
		FileDir:          "~/.config/fcli/keyring",
		FilePasswordFunc: keyring.TerminalPrompt,
	})
}

// Set stores the token for a profile, replacing any existing value.
func Set(profile, token string) error {
	kr, err := open()
	if err != nil {
		return fmt.Errorf("open keychain: %w", err)
	}
	return kr.Set(keyring.Item{
		Key:         profile,
		Data:        []byte(token),
		Label:       fmt.Sprintf("FreeSign API token (%s)", profile),
		Description: "fcli",
	})
}

// Get returns the token for a profile, or ErrNotFound.
func Get(profile string) (string, error) {
	kr, err := open()
	if err != nil {
		return "", fmt.Errorf("open keychain: %w", err)
	}
	item, err := kr.Get(profile)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("read keychain: %w", err)
	}
	return string(item.Data), nil
}

// Delete removes a profile's token. Missing tokens are a no-op.
func Delete(profile string) error {
	kr, err := open()
	if err != nil {
		return fmt.Errorf("open keychain: %w", err)
	}
	if err := kr.Remove(profile); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return fmt.Errorf("remove keychain entry: %w", err)
	}
	return nil
}
