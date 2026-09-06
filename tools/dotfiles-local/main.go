// dotfiles-local maintains ~/.config/dotfiles/local.toml: the machine-specific
// and private half of this repository's Home Manager profile — which user and
// home directory to build for, work git identities, per-host credential
// helpers, and private cln providers.
//
// The flake reads that file as its "local" input, so activation needs no
// --impure. See README.md and ../../home/local-config.md.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const usage = `usage: dotfiles-local <command> [arguments]

setup
  ensure [-f]                 record this host in [machine]; print the config directory
  init                        interactive setup for identities, helpers, and providers

inspect
  path                        print the config directory
  show                        print the current local.toml
  check                       validate local.toml without changing it

edit
  identity add <name> --author <name> --email <email> --gitdir <prefix> [--gitdir <prefix>]
  identity rm <name>
  helper set <host> <command>
  helper rm <host>
  provider add <alias> --type <github|gitlab> --host <host> [--namespace <ns>] [--default]
  provider rm <alias>
  default-provider [<alias>]

Changes take effect on the next ` + "`just switch`" + `. Set DOTFILES_LOCAL_DIR to use a
directory other than $XDG_CONFIG_HOME/dotfiles.
`

type env struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func main() {
	e := &env{in: os.Stdin, out: os.Stdout, err: os.Stderr}
	if err := run(e, os.Args[1:]); err != nil {
		fmt.Fprintf(e.err, "dotfiles-local: %v\n", err)
		os.Exit(1)
	}
}

func run(env *env, argv []string) error {
	if len(argv) == 0 {
		fmt.Fprint(env.err, usage)
		return errors.New("no command given")
	}
	switch argv[0] {
	case "-h", "--help", "help":
		fmt.Fprint(env.out, usage)
		return nil
	case "ensure":
		return runEnsure(env, argv[1:])
	case "init":
		return runInit(env, argv[1:])
	case "path":
		return runPath(env, argv[1:])
	case "show":
		return runShow(env, argv[1:])
	case "check":
		return runCheck(env, argv[1:])
	case "identity":
		return runIdentity(env, argv[1:])
	case "helper":
		return runHelper(env, argv[1:])
	case "provider":
		return runProvider(env, argv[1:])
	case "default-provider":
		return runDefaultProvider(env, argv[1:])
	default:
		return fmt.Errorf("unknown command %q; run `dotfiles-local help`", argv[0])
	}
}

func runEnsure(env *env, args []string) error {
	fs := newFlagSet("ensure", env.err)
	force := fs.Bool("f", false, "replace recorded [machine] values with this host's")
	fs.BoolVar(force, "force", false, "replace recorded [machine] values with this host's")
	if err := fs.Parse(args); err != nil {
		return err
	}

	dir, err := configDir()
	if err != nil {
		return err
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
	recorded := cfg.Machine
	changed := ensureMachine(&cfg, detected, *force)

	if !existed || changed {
		if err := commit(env, cfg); err != nil {
			return err
		}
		if existed {
			fmt.Fprintf(env.err, "dotfiles-local: updated [machine] in %s\n", path)
		} else {
			fmt.Fprintf(env.err, "dotfiles-local: created %s for %s on %s\n", path, cfg.Machine.Username, cfg.Machine.System)
		}
	} else if err := cfg.validate(); err != nil {
		return fmt.Errorf("%s is invalid:\n%w", path, err)
	}

	if !*force {
		for _, mismatch := range machineMismatches(recorded, detected) {
			fmt.Fprintf(env.err, "dotfiles-local: %s (run `dotfiles-local ensure -f` to re-detect)\n", mismatch)
		}
	}

	fmt.Fprintln(env.out, dir)
	return nil
}

func runPath(env *env, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("path takes no arguments, got %q", args[0])
	}
	dir, err := configDir()
	if err != nil {
		return err
	}
	fmt.Fprintln(env.out, dir)
	return nil
}

func runShow(env *env, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("show takes no arguments, got %q", args[0])
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s does not exist; run `dotfiles-local init`", path)
	}
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	_, err = env.out.Write(data)
	return err
}

func runCheck(env *env, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("check takes no arguments, got %q", args[0])
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	cfg, existed, err := loadConfig()
	if err != nil {
		return err
	}
	if !existed {
		return fmt.Errorf("%s does not exist; run `dotfiles-local init`", path)
	}
	if err := cfg.validate(); err != nil {
		return fmt.Errorf("%s is invalid:\n%w", path, err)
	}

	fmt.Fprintf(env.out, "%s\n", path)
	fmt.Fprintf(env.out, "  machine            %s, %s, %s\n", cfg.Machine.System, cfg.Machine.Username, cfg.Machine.HomeDirectory)
	fmt.Fprintf(env.out, "  git identities     %s\n", listOrNone(sortedKeys(cfg.identities())))
	fmt.Fprintf(env.out, "  credential helpers %s\n", listOrNone(sortedKeys(cfg.credentialHelpers())))
	fmt.Fprintf(env.out, "  cln providers      %s\n", listOrNone(sortedKeys(cfg.providers())))
	def := cfg.defaultProvider()
	if def == "" {
		def = builtinProvider
	}
	fmt.Fprintf(env.out, "  default provider   %s\n", def)
	return nil
}

func listOrNone(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	return strings.Join(items, ", ")
}
