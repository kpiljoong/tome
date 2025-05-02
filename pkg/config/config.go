package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultRemote string `yaml:"default_remote"`
}

func Load() (*Config, error) {
	path := Path()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config format: %w", err)
	}

	return &cfg, nil
}

func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("cannot resolve $HOME")
	}
	return filepath.Join(home, ".tome", "config.yaml")
}
