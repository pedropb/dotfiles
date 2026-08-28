package main

import (
	"strings"
	"testing"
)

func sampleConfig() Config {
	return Config{
		Machine: Machine{System: "x86_64-darwin", Username: "u", HomeDirectory: "/Users/u"},
		Git: &Git{
			Identities: map[string]Identity{
				"work": {Gitdir: []string{"~/src/a/", "~/src/b/"}, Name: "W", Email: "w@e"},
				"alt":  {Gitdir: []string{"~/src/c/"}, Name: "A", Email: "a@e"},
			},
			CredentialHelpers: map[string]string{"z.example.com": "!z", "a.example.com": "!a"},
		},
		Cln: &Cln{DefaultProvider: "corp", Providers: map[string]Provider{
			"corp": {Type: "gitlab", Host: "git.example.com", DefaultNamespace: "team"},
			"acme": {Type: "github", Host: "gh.example.com"},
		}},
	}
}

// The file is read by Nix and by humans, and rewritten by every edit command:
// its layout has to be a function of its contents only, never of Go's map
// iteration order, or every command would produce a spurious diff.
func TestEncodeIsDeterministicAndSorted(t *testing.T) {
	const want = `[machine]
system = 'x86_64-darwin'
username = 'u'
home_directory = '/Users/u'

[git]
[git.identities]
[git.identities.alt]
gitdir = ['~/src/c/']
name = 'A'
email = 'a@e'

[git.identities.work]
gitdir = ['~/src/a/', '~/src/b/']
name = 'W'
email = 'w@e'

[git.credential_helpers]
'a.example.com' = '!a'
'z.example.com' = '!z'

[cln]
default_provider = 'corp'

[cln.providers]
[cln.providers.acme]
type = 'github'
host = 'gh.example.com'

[cln.providers.corp]
type = 'gitlab'
host = 'git.example.com'
default_namespace = 'team'
`

	first, err := encodeConfig(sampleConfig())
	if err != nil {
		t.Fatalf("encodeConfig: %v", err)
	}
	body := strings.TrimPrefix(string(first), configHeader)
	if body != want {
		t.Errorf("unexpected encoding:\n--- got ---\n%s\n--- want ---\n%s", body, want)
	}
	for range 5 {
		again, err := encodeConfig(sampleConfig())
		if err != nil {
			t.Fatalf("encodeConfig: %v", err)
		}
		if string(again) != string(first) {
			t.Fatalf("encoding is not stable across runs")
		}
	}
}

func TestRoundTrip(t *testing.T) {
	data, err := encodeConfig(sampleConfig())
	if err != nil {
		t.Fatalf("encodeConfig: %v", err)
	}
	got, err := decodeConfig(data)
	if err != nil {
		t.Fatalf("decodeConfig: %v", err)
	}
	if got.Machine != sampleConfig().Machine {
		t.Errorf("machine: got %+v", got.Machine)
	}
	work := got.identities()["work"]
	if work.Name != "W" || work.Email != "w@e" || strings.Join(work.Gitdir, ",") != "~/src/a/,~/src/b/" {
		t.Errorf("identity work: got %+v", work)
	}
	if got.credentialHelpers()["z.example.com"] != "!z" {
		t.Errorf("credential helper: got %q", got.credentialHelpers()["z.example.com"])
	}
	if got.providers()["corp"].DefaultNamespace != "team" {
		t.Errorf("provider corp: got %+v", got.providers()["corp"])
	}
	if got.defaultProvider() != "corp" {
		t.Errorf("default provider: got %q", got.defaultProvider())
	}
}

// An unknown key is a setting that silently does nothing, which is the exact
// failure the old Nix-module local config made loud. Keep it loud.
func TestDecodeRejectsUnknownKeys(t *testing.T) {
	_, err := decodeConfig([]byte("[machine]\nsystem = 'x86_64-linux'\nhomedirectory = '/home/u'\n"))
	if err == nil {
		t.Fatal("expected an error for a misspelled key")
	}
	if !strings.Contains(err.Error(), "homedirectory") {
		t.Errorf("error should name the offending key, got: %v", err)
	}
}

func TestPruneDropsEmptiedSections(t *testing.T) {
	cfg := Config{Machine: sampleConfig().Machine, Git: &Git{Identities: map[string]Identity{}}, Cln: &Cln{Providers: map[string]Provider{}}}
	data, err := encodeConfig(cfg)
	if err != nil {
		t.Fatalf("encodeConfig: %v", err)
	}
	for _, section := range []string{"[git]", "[cln]"} {
		if strings.Contains(string(data), section) {
			t.Errorf("empty section %s should not be written:\n%s", section, data)
		}
	}
}

func TestValidate(t *testing.T) {
	valid := sampleConfig()
	if err := valid.validate(); err != nil {
		t.Fatalf("sample config should be valid: %v", err)
	}

	tests := map[string]struct {
		mutate func(*Config)
		want   string
	}{
		"empty system":        {func(c *Config) { c.Machine.System = "" }, "machine.system is empty"},
		"unknown system":      {func(c *Config) { c.Machine.System = "sparc-solaris" }, "not one of"},
		"empty username":      {func(c *Config) { c.Machine.Username = "" }, "machine.username is empty"},
		"relative home":       {func(c *Config) { c.Machine.HomeDirectory = "u" }, "not an absolute path"},
		"identity no gitdir":  {func(c *Config) { c.Git.Identities["work"] = Identity{Name: "W", Email: "w@e"} }, "no gitdir prefix"},
		"identity no email":   {func(c *Config) { c.Git.Identities["work"] = Identity{Gitdir: []string{"~/x"}, Name: "W"} }, "email is empty"},
		"bad provider type":   {func(c *Config) { c.Cln.Providers["corp"] = Provider{Type: "bitbucket", Host: "h"} }, `want "github" or "gitlab"`},
		"empty helper":        {func(c *Config) { c.Git.CredentialHelpers["a.example.com"] = "" }, "empty command"},
		"dangling default":    {func(c *Config) { c.Cln.DefaultProvider = "nope" }, "not a configured provider"},
		"builtin gh default":  {func(c *Config) { c.Cln.DefaultProvider = builtinProvider }, ""},
		"no default provider": {func(c *Config) { c.Cln.DefaultProvider = "" }, ""},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := sampleConfig()
			tc.mutate(&cfg)
			err := cfg.validate()
			if tc.want == "" {
				if err != nil {
					t.Fatalf("expected valid, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected an error mentioning %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q should mention %q", err, tc.want)
			}
		})
	}
}

// validate reports every problem at once: fixing one field per run is the
// difference between one edit and five.
func TestValidateReportsEveryProblem(t *testing.T) {
	cfg := Config{Cln: &Cln{DefaultProvider: "nope"}}
	err := cfg.validate()
	if err == nil {
		t.Fatal("expected errors")
	}
	for _, want := range []string{"machine.system", "machine.username", "machine.home_directory", "cln.default_provider"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q in:\n%v", want, err)
		}
	}
}

func TestConfigDirPrecedence(t *testing.T) {
	t.Setenv("DOTFILES_LOCAL_DIR", "/tmp/explicit")
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	dir, err := configDir()
	if err != nil || dir != "/tmp/explicit" {
		t.Fatalf("DOTFILES_LOCAL_DIR should win: %q, %v", dir, err)
	}

	t.Setenv("DOTFILES_LOCAL_DIR", "")
	dir, err = configDir()
	if err != nil || dir != "/tmp/xdg/dotfiles" {
		t.Fatalf("XDG_CONFIG_HOME should be used: %q, %v", dir, err)
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/tmp/home")
	dir, err = configDir()
	if err != nil || dir != "/tmp/home/.config/dotfiles" {
		t.Fatalf("fallback should be ~/.config/dotfiles: %q, %v", dir, err)
	}
}

func TestEnsureMachine(t *testing.T) {
	detected := Machine{System: "x86_64-linux", Username: "new", HomeDirectory: "/home/new"}

	empty := Config{}
	if !ensureMachine(&empty, detected, false) {
		t.Error("filling an empty machine should report a change")
	}
	if empty.Machine != detected {
		t.Errorf("machine not filled: %+v", empty.Machine)
	}

	recorded := Machine{System: "aarch64-darwin", Username: "old", HomeDirectory: "/Users/old"}
	kept := Config{Machine: recorded}
	if ensureMachine(&kept, detected, false) {
		t.Error("recorded values must survive a plain ensure")
	}
	if kept.Machine != recorded {
		t.Errorf("machine was overwritten: %+v", kept.Machine)
	}

	forced := Config{Machine: recorded}
	if !ensureMachine(&forced, detected, true) {
		t.Error("force should report a change")
	}
	if forced.Machine != detected {
		t.Errorf("force did not re-detect: %+v", forced.Machine)
	}
	if ensureMachine(&forced, detected, true) {
		t.Error("a second force with the same host should be a no-op")
	}

	partial := Config{Machine: Machine{Username: "old"}}
	if !ensureMachine(&partial, detected, false) {
		t.Error("missing fields should be filled")
	}
	if partial.Machine.Username != "old" || partial.Machine.System != detected.System {
		t.Errorf("only empty fields should change: %+v", partial.Machine)
	}
}

func TestMachineMismatches(t *testing.T) {
	recorded := Machine{System: "x86_64-linux", Username: "old", HomeDirectory: "/home/old"}
	detected := Machine{System: "x86_64-linux", Username: "new", HomeDirectory: "/home/old"}
	got := machineMismatches(recorded, detected)
	if len(got) != 1 || !strings.Contains(got[0], "machine.username") {
		t.Fatalf("expected one username mismatch, got %v", got)
	}
	if len(machineMismatches(Machine{}, detected)) != 0 {
		t.Error("unset fields are not mismatches")
	}
}
