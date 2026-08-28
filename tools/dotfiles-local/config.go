package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const configFileName = "local.toml"

const configHeader = `# Machine-specific and private Home Manager configuration.
#
# The dotfiles flake reads this file as its "local" input; run "just switch"
# to apply a change. Hand edits are read back faithfully, but any
# dotfiles-local command that rewrites the file normalizes its layout and
# drops comments.
#
# Schema and examples: home/local-config.md in the dotfiles repository.

`

// Machine is the identity the Home Manager profile is evaluated for. It lives
// in this file rather than being read from the environment at evaluation time,
// which is what lets the flake evaluate purely.
type Machine struct {
	System        string `toml:"system"`
	Username      string `toml:"username"`
	HomeDirectory string `toml:"home_directory"`
}

// Identity is a Git author identity applied under one or more directory
// prefixes, via Git's includeIf "gitdir:".
type Identity struct {
	Gitdir []string `toml:"gitdir"`
	Name   string   `toml:"name"`
	Email  string   `toml:"email"`
}

type Git struct {
	Identities        map[string]Identity `toml:"identities,omitempty"`
	CredentialHelpers map[string]string   `toml:"credential_helpers,omitempty"`
}

// Provider is a cln git-hosting backend.
type Provider struct {
	Type             string `toml:"type"`
	Host             string `toml:"host"`
	DefaultNamespace string `toml:"default_namespace,omitempty"`
}

type Cln struct {
	DefaultProvider string              `toml:"default_provider,omitempty"`
	Providers       map[string]Provider `toml:"providers,omitempty"`
}

// Config mirrors the option tree home/default.nix declares, so the Nix side is
// a direct mapping and this file is the only schema a user has to learn.
type Config struct {
	Machine Machine `toml:"machine"`
	Git     *Git    `toml:"git,omitempty"`
	Cln     *Cln    `toml:"cln,omitempty"`
}

// gitSection and clnSection allocate on first write so an untouched section
// never reaches the file as an empty table.
func (c *Config) gitSection() *Git {
	if c.Git == nil {
		c.Git = &Git{}
	}
	return c.Git
}

func (c *Config) clnSection() *Cln {
	if c.Cln == nil {
		c.Cln = &Cln{}
	}
	return c.Cln
}

// prune drops sections emptied by a removal, keeping the file free of stray
// headers and making save idempotent.
func (c *Config) prune() {
	if g := c.Git; g != nil {
		if len(g.Identities) == 0 {
			g.Identities = nil
		}
		if len(g.CredentialHelpers) == 0 {
			g.CredentialHelpers = nil
		}
		if g.Identities == nil && g.CredentialHelpers == nil {
			c.Git = nil
		}
	}
	if cl := c.Cln; cl != nil {
		if len(cl.Providers) == 0 {
			cl.Providers = nil
		}
		if cl.Providers == nil && cl.DefaultProvider == "" {
			c.Cln = nil
		}
	}
}

func (c Config) identities() map[string]Identity {
	if c.Git == nil {
		return nil
	}
	return c.Git.Identities
}

func (c Config) credentialHelpers() map[string]string {
	if c.Git == nil {
		return nil
	}
	return c.Git.CredentialHelpers
}

func (c Config) providers() map[string]Provider {
	if c.Cln == nil {
		return nil
	}
	return c.Cln.Providers
}

func (c Config) defaultProvider() string {
	if c.Cln == nil {
		return ""
	}
	return c.Cln.DefaultProvider
}

// configDir is $DOTFILES_LOCAL_DIR, else $XDG_CONFIG_HOME/dotfiles, else
// ~/.config/dotfiles. It is deliberately outside the repository checkout: the
// file is private, survives re-clones, and never makes the work tree dirty.
func configDir() (string, error) {
	if dir := os.Getenv("DOTFILES_LOCAL_DIR"); dir != "" {
		return dir, nil
	}
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "dotfiles"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locating home directory: %w", err)
	}
	return filepath.Join(home, ".config", "dotfiles"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// loadConfig reports whether the file exists so callers can distinguish "not
// set up yet" from "set up and empty".
func loadConfig() (Config, bool, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, false, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, false, nil
	}
	if err != nil {
		return Config{}, false, fmt.Errorf("reading %s: %w", path, err)
	}
	cfg, err := decodeConfig(data)
	if err != nil {
		return Config{}, true, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, true, nil
}

// decodeConfig rejects unknown keys: a typo such as "homedirectory" is a
// silently missing setting otherwise.
func decodeConfig(data []byte) (Config, error) {
	var cfg Config
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		var decErr *toml.DecodeError
		if errors.As(err, &decErr) {
			return Config{}, fmt.Errorf("%s\n%s", decErr, decErr.String())
		}
		var strictErr *toml.StrictMissingError
		if errors.As(err, &strictErr) {
			return Config{}, fmt.Errorf("unknown key\n%s", strictErr.String())
		}
		return Config{}, err
	}
	return cfg, nil
}

func encodeConfig(cfg Config) ([]byte, error) {
	cfg.prune()
	var buf bytes.Buffer
	buf.WriteString(configHeader)
	enc := toml.NewEncoder(&buf)
	enc.SetTablesInline(false)
	enc.SetArraysMultiline(false)
	if err := enc.Encode(cfg); err != nil {
		return nil, fmt.Errorf("encoding config: %w", err)
	}
	return buf.Bytes(), nil
}

// saveConfig writes atomically and keeps the file private: it carries employer
// hostnames and addresses, which is the whole reason it is not in the repo.
func saveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := encodeConfig(cfg)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".local.toml.*")
	if err != nil {
		return fmt.Errorf("creating temporary file in %s: %w", dir, err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing %s: %w", tmp.Name(), err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("writing %s: %w", tmp.Name(), err)
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return fmt.Errorf("setting permissions on %s: %w", tmp.Name(), err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}

// validate enforces what the Nix module's types cannot express as clearly, and
// reports every problem at once rather than one per run.
func (c Config) validate() error {
	var problems []error

	if c.Machine.System == "" {
		problems = append(problems, errors.New("machine.system is empty; run `dotfiles-local ensure`"))
	} else if !isKnownSystem(c.Machine.System) {
		problems = append(problems, fmt.Errorf("machine.system %q is not one of %s",
			c.Machine.System, strings.Join(knownSystems, ", ")))
	}
	if c.Machine.Username == "" {
		problems = append(problems, errors.New("machine.username is empty; run `dotfiles-local ensure`"))
	}
	switch {
	case c.Machine.HomeDirectory == "":
		problems = append(problems, errors.New("machine.home_directory is empty; run `dotfiles-local ensure`"))
	case !filepath.IsAbs(c.Machine.HomeDirectory):
		problems = append(problems, fmt.Errorf("machine.home_directory %q is not an absolute path", c.Machine.HomeDirectory))
	}

	for _, name := range sortedKeys(c.identities()) {
		identity := c.identities()[name]
		if len(identity.Gitdir) == 0 {
			problems = append(problems, fmt.Errorf("git.identities.%s has no gitdir prefix", name))
		}
		for _, dir := range identity.Gitdir {
			if dir == "" {
				problems = append(problems, fmt.Errorf("git.identities.%s has an empty gitdir prefix", name))
			}
		}
		if identity.Name == "" {
			problems = append(problems, fmt.Errorf("git.identities.%s.name is empty", name))
		}
		if identity.Email == "" {
			problems = append(problems, fmt.Errorf("git.identities.%s.email is empty", name))
		}
	}

	for _, host := range sortedKeys(c.credentialHelpers()) {
		if c.credentialHelpers()[host] == "" {
			problems = append(problems, fmt.Errorf("git.credential_helpers.%q has an empty command", host))
		}
	}

	for _, alias := range sortedKeys(c.providers()) {
		provider := c.providers()[alias]
		if provider.Type != "github" && provider.Type != "gitlab" {
			problems = append(problems, fmt.Errorf("cln.providers.%s.type is %q, want \"github\" or \"gitlab\"", alias, provider.Type))
		}
		if provider.Host == "" {
			problems = append(problems, fmt.Errorf("cln.providers.%s.host is empty", alias))
		}
	}

	// gh is declared by home/default.nix, not here, so it is a valid target
	// even with no [cln.providers] table at all.
	if def := c.defaultProvider(); def != "" && def != builtinProvider {
		if _, ok := c.providers()[def]; !ok {
			problems = append(problems, fmt.Errorf(
				"cln.default_provider is %q, which is not a configured provider (known: %s)",
				def, strings.Join(append([]string{builtinProvider}, sortedKeys(c.providers())...), ", ")))
		}
	}

	return errors.Join(problems...)
}

// builtinProvider is the alias home/default.nix defines unconditionally.
const builtinProvider = "gh"

var knownSystems = []string{"aarch64-darwin", "aarch64-linux", "x86_64-darwin", "x86_64-linux"}

func isKnownSystem(system string) bool {
	for _, known := range knownSystems {
		if known == system {
			return true
		}
	}
	return false
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
