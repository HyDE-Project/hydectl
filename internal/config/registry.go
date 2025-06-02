package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

// ConfigFile represents a single configuration file with its metadata and hooks
type ConfigFile struct {
	Description string   `toml:"description"`
	Path        string   `toml:"path"`
	PreHook     []string `toml:"pre_hook"`
	PostHook    []string `toml:"post_hook"`
}

// AppConfig represents an application with its files and metadata
type AppConfig struct {
	Description string                `toml:"description"`
	Icon        string                `toml:"icon"`
	Files       map[string]ConfigFile `toml:"files"`
}

// ConfigRegistry represents the entire configuration registry
type ConfigRegistry struct {
	Apps map[string]AppConfig
}

// LoadConfigRegistry loads the configuration registry from TOML file
func LoadConfigRegistry() (*ConfigRegistry, error) {
	// Look for config-registry.toml in various locations
	configPaths := []string{
		"./test/config-registry.toml", // Development path
		filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "hydectl", "config-registry.toml"),
		filepath.Join(os.Getenv("HOME"), ".config", "hydectl", "config-registry.toml"),
		filepath.Join(os.Getenv("HOME"), ".local", "lib", "hydectl", "config-registry.toml"),
		"/usr/local/lib/hydectl/config-registry.toml",
		"/usr/lib/hydectl/config-registry.toml",
	}

	// Handle XDG_CONFIG_HOME default
	if os.Getenv("XDG_CONFIG_HOME") == "" {
		configPaths[1] = filepath.Join(os.Getenv("HOME"), ".config", "hydectl", "config-registry.toml")
	}

	var configPath string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}

	if configPath == "" {
		return nil, fmt.Errorf("config-registry.toml not found in any of the expected locations")
	}

	var registry ConfigRegistry
	_, err := toml.DecodeFile(configPath, &registry.Apps)
	if err != nil {
		return nil, fmt.Errorf("error parsing config registry: %w", err)
	}

	return &registry, nil
}

// ExpandPath expands environment variables and tilde in file paths
func ExpandPath(path string) string {
	// Handle tilde expansion
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(os.Getenv("HOME"), path[2:])
	}

	// Handle environment variable expansion like ${VAR} and ${VAR:-default}
	envVarPattern := regexp.MustCompile(`\$\{([^}]+)\}`)

	expanded := envVarPattern.ReplaceAllStringFunc(path, func(match string) string {
		// Remove ${ and } to get the variable expression
		varExpr := match[2 : len(match)-1]

		// Check if it has a default value syntax VAR:-default
		if strings.Contains(varExpr, ":-") {
			parts := strings.SplitN(varExpr, ":-", 2)
			varName := parts[0]
			defaultValue := parts[1]

			if value := os.Getenv(varName); value != "" {
				return value
			}
			// Recursively expand the default value in case it contains variables
			return ExpandPath(defaultValue)
		}

		// Simple variable expansion
		return os.Getenv(varExpr)
	})

	// Handle simple $VAR syntax
	simpleVarPattern := regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	expanded = simpleVarPattern.ReplaceAllStringFunc(expanded, func(match string) string {
		varName := match[1:] // Remove the $ prefix
		return os.Getenv(varName)
	})

	return expanded
}

// FileExists checks if a configuration file exists
func (c *ConfigFile) FileExists() bool {
	expandedPath := ExpandPath(c.Path)
	_, err := os.Stat(expandedPath)
	return err == nil
}
