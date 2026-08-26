package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Provider describes a single git-hosting backend cln can clone from.
type Provider struct {
	Alias            string
	Type             string // "github" or "gitlab"
	Host             string
	DefaultNamespace string
}

// Config is cln's fully resolved configuration.
type Config struct {
	DefaultProvider string
	Providers       map[string]Provider
}

// defaultConfig is used when no config file exists: github.com only, no
// default namespace. A user (or, in this repository, the Nix module in
// home/default.nix) adds their own namespace and any private providers.
func defaultConfig() Config {
	return Config{
		DefaultProvider: "gh",
		Providers: map[string]Provider{
			"gh": {Alias: "gh", Type: "github", Host: "github.com"},
		},
	}
}

func configPath() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "cln", "config.toml")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "cln", "config.toml")
}

func loadConfig() (Config, error) {
	data, err := os.ReadFile(configPath())
	if errors.Is(err, os.ErrNotExist) {
		return defaultConfig(), nil
	}
	if err != nil {
		return Config{}, err
	}
	return parseConfig(data)
}

// parseConfig reads the minimal TOML subset cln needs: a root-level
// default_provider string key, and one level of [providers.<alias>] tables
// with the string keys type, host, and default_namespace. This is a
// deliberate subset, not a general TOML parser, so cln has zero third-party
// dependencies (see README.md). The file cln reads is still valid TOML —
// pkgs.formats.toml in home/default.nix generates it — so a real parser
// could replace this one without changing the on-disk format.
func parseConfig(data []byte) (Config, error) {
	cfg := Config{Providers: map[string]Provider{}}
	section := ""

	for i, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		lineNo := i + 1
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") {
			name := strings.TrimSpace(strings.Trim(line, "[]"))
			alias, ok := strings.CutPrefix(name, "providers.")
			if !ok {
				return Config{}, fmt.Errorf("config.toml:%d: unsupported section %q", lineNo, name)
			}
			section = strings.Trim(alias, `"`)
			if _, exists := cfg.Providers[section]; !exists {
				cfg.Providers[section] = Provider{Alias: section}
			}
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Config{}, fmt.Errorf("config.toml:%d: expected key = value", lineNo)
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"`)

		if section == "" {
			if key != "default_provider" {
				return Config{}, fmt.Errorf("config.toml:%d: unsupported key %q", lineNo, key)
			}
			cfg.DefaultProvider = value
			continue
		}

		p := cfg.Providers[section]
		switch key {
		case "type":
			p.Type = value
		case "host":
			p.Host = value
		case "default_namespace":
			p.DefaultNamespace = value
		default:
			return Config{}, fmt.Errorf("config.toml:%d: unsupported key %q", lineNo, key)
		}
		cfg.Providers[section] = p
	}

	return cfg, cfg.validate()
}

func (cfg Config) validate() error {
	if cfg.DefaultProvider == "" {
		return errors.New("config.toml: default_provider is required")
	}
	if _, ok := cfg.Providers[cfg.DefaultProvider]; !ok {
		return fmt.Errorf("config.toml: default_provider %q is not a configured provider", cfg.DefaultProvider)
	}
	for alias, p := range cfg.Providers {
		if p.Type != "github" && p.Type != "gitlab" {
			return fmt.Errorf("config.toml: providers.%s: unsupported type %q (want github or gitlab)", alias, p.Type)
		}
		if p.Host == "" {
			return fmt.Errorf("config.toml: providers.%s: host is required", alias)
		}
	}
	return nil
}
