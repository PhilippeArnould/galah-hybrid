package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the configuration file settings for the application.
type Config struct {
	SystemPrompt string               `yaml:"system_prompt"`
	UserPrompt   string               `yaml:"user_prompt"`
	Ports        []PortConfig         `yaml:"ports"`
	Profiles     map[string]TLSConfig `yaml:"profiles"`
	Scenario     ScenarioConfig       `yaml:"scenario"`
}

// TLSConfig contains TLS-related settings.
type TLSConfig struct {
	Certificate string `yaml:"certificate"`
	Key         string `yaml:"key"`
}

// PortConfig specifies honeypot port settings.
type PortConfig struct {
	Port       uint16 `yaml:"port"`
	Protocol   string `yaml:"protocol"`
	TLSProfile string `yaml:"tls_profile"`
}

// LoadConfig reads and parses the configuration file.
func LoadConfig(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ScenarioConfig selects an optional set of HTML templates. Routes are ordered.
type ScenarioConfig struct {
	Enabled      bool          `yaml:"enabled"`
	Name         string        `yaml:"name"`
	TemplatesDir string        `yaml:"templates_dir"`
	Routes       []RouteConfig `yaml:"routes"`
}

type RouteConfig struct {
	Pattern  string   `yaml:"pattern"`
	Template string   `yaml:"template"`
	Fields   []string `yaml:"fields"`
}
