package main

import (
	"errors"
	"fmt"
	"os"
)

// runInit is the first-run entry point `just bootstrap` calls: it detects
// this machine, prefills [machine], and hands off to the same full-screen
// editor `edit` uses for everything else.
func runInit(env *env, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("init takes no arguments, got %q", args[0])
	}
	if !interactive() {
		return errors.New("init needs an interactive terminal; use the identity, helper, and provider commands instead")
	}

	path, err := configPath()
	if err != nil {
		return err
	}
	cfg, existed, err := loadConfig()
	if err != nil {
		return err
	}
	detected, err := detectedMachine()
	if err != nil {
		return err
	}
	ensureMachine(&cfg, detected, false)

	return runEditor(env, cfg, path, existed)
}

// runEdit opens the same full-screen editor as `init`, for touching up an
// existing (or brand new) local.toml without redoing the whole wizard.
func runEdit(env *env, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("edit takes no arguments, got %q", args[0])
	}
	if !interactive() {
		return errors.New("edit needs an interactive terminal; use the identity, helper, and provider commands instead")
	}

	path, err := configPath()
	if err != nil {
		return err
	}
	cfg, existed, err := loadConfig()
	if err != nil {
		return err
	}
	detected, err := detectedMachine()
	if err != nil {
		return err
	}
	ensureMachine(&cfg, detected, false)

	return runEditor(env, cfg, path, existed)
}

// interactive reports whether stdin is a real terminal. A full-screen editor
// needs one; over a non-interactive stdin (a script, CI, an SSH command with
// no pty) it would hang or crash instead of failing cleanly.
func interactive() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
