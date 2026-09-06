package main

import (
	"path/filepath"
	"testing"
)

func TestDefaultDest(t *testing.T) {
	t.Setenv("HOME", "/home/pedropb")
	got := defaultDest("github.com", "pedropb", "dotfiles")
	want := filepath.Join("/home/pedropb", "src", "github.com", "pedropb", "dotfiles")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func testProviders() map[string]Provider {
	return map[string]Provider{
		"gh":   {Alias: "gh", Type: "github", Host: "github.com", DefaultNamespace: "pedropb"},
		"corp": {Alias: "corp", Type: "gitlab", Host: "git.example.com"},
	}
}

func TestParseArgsRepoOnly(t *testing.T) {
	sp, err := parseArgs([]string{"dotfiles"}, testProviders())
	if err != nil {
		t.Fatal(err)
	}
	if sp.providerAlias != "" || sp.repoPath != "dotfiles" || sp.dest != "" {
		t.Errorf("got %+v", sp)
	}
}

func TestParseArgsProviderAndRepo(t *testing.T) {
	sp, err := parseArgs([]string{"gh", "dotfiles"}, testProviders())
	if err != nil {
		t.Fatal(err)
	}
	if sp.providerAlias != "gh" || sp.repoPath != "dotfiles" || sp.dest != "" {
		t.Errorf("got %+v", sp)
	}
}

func TestParseArgsRepoAndDest(t *testing.T) {
	sp, err := parseArgs([]string{"team/billing", "/tmp/x"}, testProviders())
	if err != nil {
		t.Fatal(err)
	}
	if sp.providerAlias != "" || sp.repoPath != "team/billing" || sp.dest != "/tmp/x" {
		t.Errorf("got %+v", sp)
	}
}

func TestParseArgsProviderRepoAndDest(t *testing.T) {
	sp, err := parseArgs([]string{"corp", "team/billing", "/tmp/x"}, testProviders())
	if err != nil {
		t.Fatal(err)
	}
	if sp.providerAlias != "corp" || sp.repoPath != "team/billing" || sp.dest != "/tmp/x" {
		t.Errorf("got %+v", sp)
	}
}

func TestParseArgsUnknownProviderWithThreeArgs(t *testing.T) {
	if _, err := parseArgs([]string{"nope", "team/billing", "/tmp/x"}, testProviders()); err == nil {
		t.Fatal("expected an unknown-provider error")
	}
}

func TestSplitRepoPathUsesDefaultNamespace(t *testing.T) {
	gh := testProviders()["gh"]
	namespace, repo, err := splitRepoPath("dotfiles", gh)
	if err != nil {
		t.Fatal(err)
	}
	if namespace != "pedropb" || repo != "dotfiles" {
		t.Errorf("got namespace=%q repo=%q", namespace, repo)
	}
}

func TestSplitRepoPathRequiresNamespaceWhenNoDefault(t *testing.T) {
	corp := testProviders()["corp"]
	if _, _, err := splitRepoPath("billing", corp); err == nil {
		t.Fatal("expected an error: corp has no default namespace")
	}
}

func TestSplitRepoPathRejectsNestedGitHubNamespace(t *testing.T) {
	gh := testProviders()["gh"]
	if _, _, err := splitRepoPath("team/sub/repo", gh); err == nil {
		t.Fatal("expected an error: github namespaces cannot nest")
	}
}

func TestSplitRepoPathAllowsNestedGitLabNamespace(t *testing.T) {
	corp := testProviders()["corp"]
	namespace, repo, err := splitRepoPath("team/sub/repo", corp)
	if err != nil {
		t.Fatal(err)
	}
	if namespace != "team/sub" || repo != "repo" {
		t.Errorf("got namespace=%q repo=%q", namespace, repo)
	}
}

func TestCloneURL(t *testing.T) {
	gh := testProviders()["gh"]
	got := cloneURL(gh, "pedropb", "dotfiles")
	want := "https://github.com/pedropb/dotfiles"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
