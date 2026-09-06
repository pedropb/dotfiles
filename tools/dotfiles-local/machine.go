package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// detectedMachine describes the host dotfiles-local is running on. Recording
// it into local.toml is the whole trick behind pure evaluation: the ambient
// environment is read once, here, and written down, instead of being read by
// builtins.getEnv on every `nix build`.
func detectedMachine() (Machine, error) {
	system, err := nixSystem()
	if err != nil {
		return Machine{}, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return Machine{}, fmt.Errorf("locating home directory: %w", err)
	}
	return Machine{
		System:        system,
		Username:      detectUsername(home),
		HomeDirectory: home,
	}, nil
}

// nixSystem maps Go's platform names onto Nix's double. Deriving it in-process
// keeps `ensure` working before Nix is on PATH and avoids a subprocess.
func nixSystem() (string, error) {
	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		return "", fmt.Errorf("unsupported architecture %q", runtime.GOARCH)
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		return arch + "-" + runtime.GOOS, nil
	default:
		return "", fmt.Errorf("unsupported operating system %q", runtime.GOOS)
	}
}

// detectUsername prefers $USER: os/user falls back to parsing /etc/passwd in
// CGO-less builds, which does not list ordinary accounts on macOS.
func detectUsername(home string) string {
	if name := strings.TrimSpace(os.Getenv("USER")); name != "" {
		return name
	}
	if current, err := user.Current(); err == nil && current.Username != "" {
		return current.Username
	}
	return filepath.Base(home)
}

func ensureMachine(cfg *Config, detected Machine, force bool) bool {
	changed := false
	set := func(field *string, value string) {
		if value == "" || (*field == value) || (*field != "" && !force) {
			return
		}
		*field = value
		changed = true
	}
	set(&cfg.Machine.System, detected.System)
	set(&cfg.Machine.Username, detected.Username)
	set(&cfg.Machine.HomeDirectory, detected.HomeDirectory)
	return changed
}

func machineMismatches(recorded, detected Machine) []string {
	var out []string
	compare := func(key, have, want string) {
		if have != "" && want != "" && have != want {
			out = append(out, fmt.Sprintf("machine.%s is %q but this host is %q", key, have, want))
		}
	}
	compare("system", recorded.System, detected.System)
	compare("username", recorded.Username, detected.Username)
	compare("home_directory", recorded.HomeDirectory, detected.HomeDirectory)
	return out
}
