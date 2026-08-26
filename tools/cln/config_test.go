package main

import "testing"

func TestParseConfig(t *testing.T) {
	data := []byte(`
default_provider = "gh"

[providers.gh]
type = "github"
host = "github.com"
default_namespace = "pedropb"

[providers.corp]
type = "gitlab"
host = "git.example.com"
`)

	cfg, err := parseConfig(data)
	if err != nil {
		t.Fatalf("parseConfig: %v", err)
	}
	if cfg.DefaultProvider != "gh" {
		t.Errorf("default_provider = %q, want gh", cfg.DefaultProvider)
	}

	gh := cfg.Providers["gh"]
	if gh.Type != "github" || gh.Host != "github.com" || gh.DefaultNamespace != "pedropb" {
		t.Errorf("providers.gh = %+v", gh)
	}

	corp := cfg.Providers["corp"]
	if corp.Type != "gitlab" || corp.Host != "git.example.com" || corp.DefaultNamespace != "" {
		t.Errorf("providers.corp = %+v", corp)
	}
}

func TestParseConfigRejectsUnknownType(t *testing.T) {
	data := []byte(`
default_provider = "x"

[providers.x]
type = "bitbucket"
host = "bitbucket.org"
`)
	if _, err := parseConfig(data); err == nil {
		t.Fatal("expected error for an unsupported provider type")
	}
}

func TestParseConfigRequiresDefaultProvider(t *testing.T) {
	data := []byte(`
[providers.gh]
type = "github"
host = "github.com"
`)
	if _, err := parseConfig(data); err == nil {
		t.Fatal("expected error when default_provider is missing")
	}
}

func TestParseConfigRequiresKnownDefaultProvider(t *testing.T) {
	data := []byte(`
default_provider = "missing"

[providers.gh]
type = "github"
host = "github.com"
`)
	if _, err := parseConfig(data); err == nil {
		t.Fatal("expected error when default_provider names an undeclared provider")
	}
}

func TestDefaultConfigIsValid(t *testing.T) {
	if err := defaultConfig().validate(); err != nil {
		t.Fatalf("defaultConfig() is invalid: %v", err)
	}
}
