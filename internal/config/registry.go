package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

type ConfigFile struct {
	Description string   `toml:"description"`
	Path        string   `toml:"path"`
	PreHook     []string `toml:"pre_hook"`
	PostHook    []string `toml:"post_hook"`
}

type AppConfig struct {
	Description string                `toml:"description"`
	Icon        string                `toml:"icon"`
	Files       map[string]ConfigFile `toml:"files"`
}

type ConfigRegistry struct {
	Apps map[string]AppConfig
}

func LoadConfigRegistry() (*ConfigRegistry, error) {

	configPaths := []string{
		"./test/config-registry.toml",
		filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "hydectl", "config-registry.toml"),
		filepath.Join(os.Getenv("HOME"), ".config", "hydectl", "config-registry.toml"),
		filepath.Join(os.Getenv("HOME"), ".local", "lib", "hydectl", "config-registry.toml"),
		"/usr/local/lib/hydectl/config-registry.toml",
		"/usr/lib/hydectl/config-registry.toml",
	}

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

func ExpandPath(path string) string {

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(os.Getenv("HOME"), path[2:])
	}

	envVarPattern := regexp.MustCompile(`\$\{([^}]+)\}`)

	expanded := envVarPattern.ReplaceAllStringFunc(path, func(match string) string {

		varExpr := match[2 : len(match)-1]

		if strings.Contains(varExpr, ":-") {
			parts := strings.SplitN(varExpr, ":-", 2)
			varName := parts[0]
			defaultValue := parts[1]

			if value := os.Getenv(varName); value != "" {
				return value
			}

			return ExpandPath(defaultValue)
		}

		return os.Getenv(varExpr)
	})

	simpleVarPattern := regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	expanded = simpleVarPattern.ReplaceAllStringFunc(expanded, func(match string) string {
		varName := match[1:]
		return os.Getenv(varName)
	})

	return expanded
}

func (c *ConfigFile) FileExists() bool {
	expandedPath := ExpandPath(c.Path)
	_, err := os.Stat(expandedPath)
	return err == nil
}
