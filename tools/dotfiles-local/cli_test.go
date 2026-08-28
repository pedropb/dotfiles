package main

import (
	"bytes"
	"strings"
	"testing"
)

type harness struct {
	t   *testing.T
	dir string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("DOTFILES_LOCAL_DIR", dir)
	t.Setenv("HOME", dir)
	t.Setenv("USER", "tester")
	return &harness{t: t, dir: dir}
}

func (h *harness) run(args ...string) (string, string, error) {
	h.t.Helper()
	var out, errOut bytes.Buffer
	err := run(&env{in: strings.NewReader(""), out: &out, err: &errOut}, args)
	return out.String(), errOut.String(), err
}

func (h *harness) mustRun(args ...string) string {
	h.t.Helper()
	out, errOut, err := h.run(args...)
	if err != nil {
		h.t.Fatalf("%v failed: %v\nstderr: %s", args, err, errOut)
	}
	return out
}

func (h *harness) load() Config {
	h.t.Helper()
	cfg, existed, err := loadConfig()
	if err != nil {
		h.t.Fatalf("loading config: %v", err)
	}
	if !existed {
		h.t.Fatal("config file does not exist")
	}
	return cfg
}

// `just switch` consumes ensure's stdout as the --override-input path, so it
// must be the bare directory and nothing else, even on the run that creates
// the file and logs about it.
func TestEnsureCreatesConfigAndPrintsOnlyTheDirectory(t *testing.T) {
	h := newHarness(t)

	out, errOut, err := h.run("ensure")
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if got := strings.TrimSpace(out); got != h.dir {
		t.Errorf("stdout should be the config directory alone, got %q", out)
	}
	if !strings.Contains(errOut, "created") {
		t.Errorf("first ensure should say it created the file, got %q", errOut)
	}

	cfg := h.load()
	if cfg.Machine.Username != "tester" || cfg.Machine.HomeDirectory != h.dir {
		t.Errorf("machine not detected: %+v", cfg.Machine)
	}
	if !isKnownSystem(cfg.Machine.System) {
		t.Errorf("machine.system %q is not a Nix double", cfg.Machine.System)
	}

	out, errOut, err = h.run("ensure")
	if err != nil {
		t.Fatalf("second ensure: %v", err)
	}
	if strings.TrimSpace(out) != h.dir {
		t.Errorf("stdout changed on the second run: %q", out)
	}
	if errOut != "" {
		t.Errorf("a settled config should make ensure silent, got %q", errOut)
	}
}

func TestEnsureWarnsOnMismatchAndForceRewrites(t *testing.T) {
	h := newHarness(t)
	h.mustRun("ensure")

	t.Setenv("USER", "someone-else")
	_, errOut, err := h.run("ensure")
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if !strings.Contains(errOut, "machine.username") || !strings.Contains(errOut, "ensure -f") {
		t.Errorf("mismatch should be reported with the fix, got %q", errOut)
	}
	if h.load().Machine.Username != "tester" {
		t.Error("a plain ensure must not rewrite a recorded username")
	}

	h.mustRun("ensure", "-f")
	if got := h.load().Machine.Username; got != "someone-else" {
		t.Errorf("ensure -f should re-detect, got %q", got)
	}
}

func TestIdentityLifecycle(t *testing.T) {
	h := newHarness(t)
	h.mustRun("ensure")

	h.mustRun("identity", "add", "work",
		"--author", "Example Author", "--email", "author@example.com",
		"--gitdir", "~/src/git.example.com/", "--gitdir", "~/src/example.com/")

	identity := h.load().identities()["work"]
	if identity.Name != "Example Author" || identity.Email != "author@example.com" {
		t.Fatalf("identity not stored: %+v", identity)
	}
	if strings.Join(identity.Gitdir, ",") != "~/src/git.example.com/,~/src/example.com/" {
		t.Fatalf("repeated --gitdir should accumulate in order: %+v", identity.Gitdir)
	}

	h.mustRun("identity", "rm", "work")
	if _, ok := h.load().identities()["work"]; ok {
		t.Error("identity should be gone")
	}
	if _, _, err := h.run("identity", "rm", "work"); err == nil {
		t.Error("removing a missing identity should fail")
	}
}

func TestIdentityAddRequiresEveryField(t *testing.T) {
	h := newHarness(t)
	h.mustRun("ensure")

	if _, _, err := h.run("identity", "add", "work", "--author", "A", "--email", "a@e"); err == nil {
		t.Error("an identity with no gitdir should be rejected")
	}
	if _, _, err := h.run("identity", "add", "--author", "A"); err == nil {
		t.Error("a missing positional name should be rejected")
	}
	if h.load().Git != nil {
		t.Error("a rejected edit must not reach the file")
	}
}

func TestHelperLifecycle(t *testing.T) {
	h := newHarness(t)
	h.mustRun("ensure")

	h.mustRun("helper", "set", "git.example.com", "!glab auth git-credential")
	if got := h.load().credentialHelpers()["git.example.com"]; got != "!glab auth git-credential" {
		t.Fatalf("helper not stored: %q", got)
	}

	h.mustRun("helper", "rm", "git.example.com")
	if len(h.load().credentialHelpers()) != 0 {
		t.Error("helper should be gone")
	}
	if _, _, err := h.run("helper", "rm", "git.example.com"); err == nil {
		t.Error("removing a missing helper should fail")
	}
}

func TestProviderLifecycleKeepsDefaultProviderValid(t *testing.T) {
	h := newHarness(t)
	h.mustRun("ensure")

	h.mustRun("provider", "add", "corp", "--type", "gitlab", "--host", "git.example.com", "--namespace", "team", "--default")
	cfg := h.load()
	if provider := cfg.providers()["corp"]; provider.Type != "gitlab" || provider.Host != "git.example.com" || provider.DefaultNamespace != "team" {
		t.Fatalf("provider not stored: %+v", provider)
	}
	if cfg.defaultProvider() != "corp" {
		t.Fatalf("--default should set the default provider, got %q", cfg.defaultProvider())
	}

	// Removing the default must not leave a dangling alias behind: the next
	// `just switch` would fail the module assertion instead.
	h.mustRun("provider", "rm", "corp")
	cfg = h.load()
	if len(cfg.providers()) != 0 {
		t.Error("provider should be gone")
	}
	if cfg.defaultProvider() != "" {
		t.Errorf("default provider should fall back to the built-in, got %q", cfg.defaultProvider())
	}
	if err := cfg.validate(); err != nil {
		t.Errorf("config should still be valid: %v", err)
	}
}

func TestProviderAddRejectsUnknownType(t *testing.T) {
	h := newHarness(t)
	h.mustRun("ensure")
	if _, _, err := h.run("provider", "add", "corp", "--type", "bitbucket", "--host", "h"); err == nil {
		t.Fatal("an unsupported backend should be rejected")
	}
	if h.load().Cln != nil {
		t.Error("a rejected provider must not reach the file")
	}
}

func TestDefaultProviderCommand(t *testing.T) {
	h := newHarness(t)
	h.mustRun("ensure")
	h.mustRun("provider", "add", "corp", "--type", "gitlab", "--host", "git.example.com")

	if got := strings.TrimSpace(h.mustRun("default-provider")); got != builtinProvider+" (default)" {
		t.Errorf("unset default should report the built-in, got %q", got)
	}
	h.mustRun("default-provider", "corp")
	if h.load().defaultProvider() != "corp" {
		t.Error("default provider not set")
	}
	h.mustRun("default-provider", builtinProvider)
	if got := h.load().defaultProvider(); got != "" {
		t.Errorf("selecting the built-in should clear the key, got %q", got)
	}
	if _, _, err := h.run("default-provider", "nope"); err == nil {
		t.Error("an unknown alias should be rejected")
	}
}

func TestCheckReportsInvalidFileAndSummary(t *testing.T) {
	h := newHarness(t)

	if _, _, err := h.run("check"); err == nil {
		t.Error("check should fail before the config exists")
	}

	h.mustRun("ensure")
	h.mustRun("provider", "add", "corp", "--type", "gitlab", "--host", "git.example.com", "--default")
	out := h.mustRun("check")
	for _, want := range []string{"machine", "cln providers", "corp", "default provider"} {
		if !strings.Contains(out, want) {
			t.Errorf("check output should mention %q:\n%s", want, out)
		}
	}
}

func TestShowAndPath(t *testing.T) {
	h := newHarness(t)
	if _, _, err := h.run("show"); err == nil {
		t.Error("show should fail before the config exists")
	}
	if got := strings.TrimSpace(h.mustRun("path")); got != h.dir {
		t.Errorf("path should print the config directory, got %q", got)
	}

	h.mustRun("ensure")
	if out := h.mustRun("show"); !strings.Contains(out, "[machine]") {
		t.Errorf("show should print the file, got %q", out)
	}
}

func TestUnknownCommand(t *testing.T) {
	h := newHarness(t)
	_, _, err := h.run("frobnicate")
	if err == nil || !strings.Contains(err.Error(), "frobnicate") {
		t.Errorf("expected an error naming the command, got %v", err)
	}
}
