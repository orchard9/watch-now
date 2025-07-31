package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Services []ServiceConfig `yaml:"services"`
	Checks   []CheckConfig   `yaml:"checks"`
	Interval time.Duration   `yaml:"interval"`
	API      APIConfig       `yaml:"api"`
}

type ServiceConfig struct {
	Name    string            `yaml:"name"`
	Type    string            `yaml:"type"`
	URL     string            `yaml:"url"`
	Health  string            `yaml:"health"`
	Headers map[string]string `yaml:"headers"`
	Timeout time.Duration     `yaml:"timeout"`
}

type CheckConfig struct {
	Name    string        `yaml:"name"`
	Command string        `yaml:"command"`
	Args    []string      `yaml:"args"`
	Timeout time.Duration `yaml:"timeout"`
}

type APIConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Set defaults
	if config.Interval == 0 {
		config.Interval = 60 * time.Second
	}
	if config.API.Port == 0 {
		config.API.Port = 0 // Use ephemeral port
	}

	// Set default timeouts
	for i := range config.Services {
		if config.Services[i].Timeout == 0 {
			config.Services[i].Timeout = 10 * time.Second
		}
	}
	for i := range config.Checks {
		if config.Checks[i].Timeout == 0 {
			config.Checks[i].Timeout = 30 * time.Second
		}
	}

	return &config, nil
}
