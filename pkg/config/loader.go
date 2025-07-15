package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from a treex.yaml file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return LoadConfigFromReader(file)
}

// LoadConfigFromReader loads configuration from an io.Reader
func LoadConfigFromReader(reader io.Reader) (*Config, error) {
	// Start with empty config to detect missing fields
	config := &Config{}

	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true) // Strict parsing

	if err := decoder.Decode(config); err != nil {
		if err == io.EOF {
			// Empty file, return default config
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Set defaults for missing fields after validation
	if config.Styles == nil {
		config.Styles = &StylesConfig{
			Theme: "auto",
		}
	} else if config.Styles.Theme == "" {
		config.Styles.Theme = "auto"
	}

	return config, nil
}

// FindConfigFile searches for a treex.yaml file in common locations
func FindConfigFile() (string, error) {
	// Check current directory first
	if fileExists("treex.yaml") {
		return "treex.yaml", nil
	}

	// Check home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".config", "treex", "treex.yaml")
		if fileExists(configPath) {
			return configPath, nil
		}

		// Also check ~/.treex.yaml
		configPath = filepath.Join(homeDir, ".treex.yaml")
		if fileExists(configPath) {
			return configPath, nil
		}
	}

	// Check XDG_CONFIG_HOME if set
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		configPath := filepath.Join(xdgConfig, "treex", "treex.yaml")
		if fileExists(configPath) {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("no treex.yaml found")
}

// LoadConfigFromDefaultLocations tries to load config from default locations
func LoadConfigFromDefaultLocations() (*Config, error) {
	configPath, err := FindConfigFile()
	if err != nil {
		// No config file found, use defaults
		return DefaultConfig(), nil
	}

	return LoadConfig(configPath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func validateConfig(config *Config) error {
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	if config.Styles != nil {
		if config.Styles.Theme != "" && config.Styles.Theme != "dark" && config.Styles.Theme != "light" && config.Styles.Theme != "auto" {
			return fmt.Errorf("invalid theme: %s (must be 'dark', 'light', or 'auto')", config.Styles.Theme)
		}
	}

	return nil
}