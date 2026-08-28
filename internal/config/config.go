// Package config resolves project/location/output settings with precedence:
// command-line flag > environment variable > ~/.castctl/config.yaml.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Settings holds resolved runtime configuration for a command invocation.
type Settings struct {
	Project  string
	Location string
	JSON     bool
	Async    bool
}

const (
	keyProject  = "project"
	keyLocation = "location"
)

// Dir returns the config directory (~/.castctl), creating nothing.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".castctl"), nil
}

// FilePath returns the config file path (~/.castctl/config.yaml).
func FilePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// newViper builds a viper instance bound to env + config file.
// Env vars: GOOGLE_CLOUD_PROJECT (and CLOUDSDK_CORE_PROJECT) for project,
// CASTCTL_LOCATION for location.
func newViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	if dir, err := Dir(); err == nil {
		v.AddConfigPath(dir)
	}
	// Best-effort read; absence of file is fine.
	_ = v.ReadInConfig()

	v.BindEnv(keyProject, "GOOGLE_CLOUD_PROJECT", "CLOUDSDK_CORE_PROJECT")
	v.BindEnv(keyLocation, "CASTCTL_LOCATION")
	return v
}

// Resolve merges flag values (highest precedence) over env and file.
// Empty flag strings mean "not set on the command line".
func Resolve(flagProject, flagLocation string, jsonOut, async bool) (Settings, error) {
	v := newViper()

	project := firstNonEmpty(flagProject, v.GetString(keyProject))
	location := firstNonEmpty(flagLocation, v.GetString(keyLocation))

	return Settings{
		Project:  project,
		Location: location,
		JSON:     jsonOut,
		Async:    async,
	}, nil
}

// RequireProjectLocation returns an error naming what is missing.
func (s Settings) RequireProjectLocation() error {
	if s.Project == "" {
		return fmt.Errorf("no project set: pass --project, set GOOGLE_CLOUD_PROJECT, or run 'castctl config set project <id>'")
	}
	if s.Location == "" {
		return fmt.Errorf("no location set: pass --location, set CASTCTL_LOCATION, or run 'castctl config set location <region>'")
	}
	return nil
}

// Set writes a single key to the config file, creating the dir/file as needed.
func Set(key, value string) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	v := viper.New()
	v.SetConfigFile(filepath.Join(dir, "config.yaml"))
	_ = v.ReadInConfig()
	v.Set(key, value)
	return v.WriteConfig()
}

// Get returns a single stored key (file only, no env/flag).
func Get(key string) (string, error) {
	v := newViper()
	return v.GetString(key), nil
}

// All returns every stored key/value pair from the config file.
func All() (map[string]any, error) {
	v := newViper()
	return v.AllSettings(), nil
}

func firstNonEmpty(vals ...string) string {
	for _, s := range vals {
		if s != "" {
			return s
		}
	}
	return ""
}
