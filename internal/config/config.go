package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Hosts  []HostConfig `yaml:"hosts"`
}

type ServerConfig struct {
	ChatIDs []int64
	Port    string `yaml:"port"`
	Host    string `yaml:"host"`
}

type HostConfig struct {
	Name      string            `yaml:"name"`
	Templates map[string]string `yaml:"templates"`
	Enabled   bool              `yaml:"enabled"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func (c *Config) GetHostConfig(hostName string) *HostConfig {
	for _, host := range c.Hosts {
		if host.Name == hostName && host.Enabled {
			return &host
		}
	}
	return nil
}
