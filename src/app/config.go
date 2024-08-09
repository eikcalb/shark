package app

import (
	"eikcalb.dev/shark/src/store"
)

/*
Config represents the application configuration.
*/
type Config struct {
	Name    string `json:"name"`
	Version string `json:"version"`

	Port uint16 `json:"port"`

	storage *store.JSONFileStore[Config] `json:"-"`
}

// Save serializes the active config and persists it.
func (c Config) Save() error {
	err := c.storage.Save(c)
	if err != nil {
		// Failed to save config.
		return err
	}

	return nil
}

// LoadConfig reads the config from JSON and returns an instance
// of the Config struct.
func LoadConfig(path string) (*Config, error) {
	jfs := store.JSONFileStore[Config]{Path: path}
	config, err := jfs.Load()
	if err != nil {
		// Failed to load config
		return nil, err
	}

	config.storage = &jfs

	return config, nil
}
