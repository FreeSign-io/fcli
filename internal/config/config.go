// Package config manages fcli's user-facing configuration.
//
// Layout:
//
//	~/.config/fcli/config.toml      # profiles + active profile pointer
//	(API tokens go in the OS keychain — see internal/keychain)
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	defaultProfile = "default"
	defaultBaseURL = "https://freesign.io"
)

// Profile holds per-instance settings (one per FreeSign deployment a user talks to).
type Profile struct {
	Name    string `mapstructure:"name"`
	BaseURL string `mapstructure:"base_url"`
}

// Config is the on-disk schema.
type Config struct {
	ActiveProfile string             `mapstructure:"active_profile"`
	Profiles      map[string]Profile `mapstructure:"profiles"`
}

// Path returns the full path to the config file (creating the parent dir if missing).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home dir: %w", err)
	}
	dir := filepath.Join(home, ".config", "fcli")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads config from disk. A missing file is not an error; we return a Config
// pre-populated with a single "default" profile pointing at https://freesign.io.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")

	if _, err := os.Stat(path); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
	}

	cfg := &Config{Profiles: map[string]Profile{}}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	if _, ok := cfg.Profiles[defaultProfile]; !ok {
		cfg.Profiles[defaultProfile] = Profile{Name: defaultProfile, BaseURL: defaultBaseURL}
	}
	if cfg.ActiveProfile == "" {
		cfg.ActiveProfile = defaultProfile
	}
	return cfg, nil
}

// Save writes the config back to disk atomically (write-temp-then-rename).
func (c *Config) Save() error {
	path, err := Path()
	if err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigType("toml")
	v.Set("active_profile", c.ActiveProfile)
	v.Set("profiles", c.Profiles)

	tmp := path + ".tmp"
	if err := v.WriteConfigAs(tmp); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		return fmt.Errorf("chmod %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", tmp, path, err)
	}
	return nil
}

// Active returns the currently selected profile.
func (c *Config) Active() Profile {
	if p, ok := c.Profiles[c.ActiveProfile]; ok {
		return p
	}
	return Profile{Name: defaultProfile, BaseURL: defaultBaseURL}
}
