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

type OrderedConfigRegistry struct {
	AppsOrder []string
	Apps      map[string]AppConfig
}

func LoadConfigRegistry() (*OrderedConfigRegistry, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}

	configPaths := []string{
		filepath.Join(configHome, "hyde", "config-registry.toml"),
		filepath.Join(dataHome, "hyde", "config-registry.toml"),
		"/usr/local/share/hyde/config-registry.toml",
		"/usr/share/hyde/config-registry.toml",
	}

	var configPath string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}

	if configPath == "" {
		return nil, fmt.Errorf("config-registry.toml not found in any of the expected locations: %v", configPaths)
	}

	var (
		appsOrder []string
		apps      = make(map[string]AppConfig)
	)

	var meta toml.MetaData
	meta, err := toml.DecodeFile(configPath, &apps)
	if err != nil {
		return nil, fmt.Errorf("error parsing config registry: %w", err)
	}

	// Normalize all keys to lower case
	normApps := make(map[string]AppConfig)
	for k, v := range apps {
		normApps[strings.ToLower(k)] = v
	}
	for _, key := range meta.Keys() {
		if len(key) == 1 {
			k := key[0]
			if _, ok := apps[k]; ok {
				appsOrder = append(appsOrder, strings.ToLower(k))
			}
		}
	}

	return &OrderedConfigRegistry{
		AppsOrder: appsOrder,
		Apps:      normApps,
	}, nil
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
