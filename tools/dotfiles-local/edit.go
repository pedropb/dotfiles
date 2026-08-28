package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// stringList collects a flag that may be repeated, e.g. --gitdir.
type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }

func (s *stringList) Set(value string) error {
	if value == "" {
		return errors.New("value is empty")
	}
	*s = append(*s, value)
	return nil
}

func newFlagSet(name string, out io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(out)
	return fs
}

// takeArg splits the leading positional argument from the flags that follow,
// because Go's flag package stops parsing at the first non-flag argument.
func takeArg(args []string, what string) (string, []string, error) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return "", nil, fmt.Errorf("missing <%s>", what)
	}
	return args[0], args[1:], nil
}

func runIdentity(env *env, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: dotfiles-local identity <add|rm> <name> ...")
	}
	switch args[0] {
	case "add":
		return identityAdd(env, args[1:])
	case "rm":
		return identityRemove(env, args[1:])
	default:
		return fmt.Errorf("unknown identity command %q", args[0])
	}
}

func identityAdd(env *env, args []string) error {
	name, rest, err := takeArg(args, "name")
	if err != nil {
		return fmt.Errorf("%w\nusage: dotfiles-local identity add <name> --author <name> --email <email> --gitdir <prefix>", err)
	}
	fs := newFlagSet("identity add", env.err)
	var gitdirs stringList
	author := fs.String("author", "", "git author name")
	email := fs.String("email", "", "git author email")
	fs.Var(&gitdirs, "gitdir", "directory prefix the identity applies to (repeatable)")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if *author == "" || *email == "" || len(gitdirs) == 0 {
		return errors.New("identity add requires --author, --email, and at least one --gitdir")
	}

	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg.setIdentity(name, Identity{Gitdir: gitdirs, Name: *author, Email: *email})
	if err := commit(env, cfg); err != nil {
		return err
	}
	fmt.Fprintf(env.out, "identity %s: %s <%s> under %s\n", name, *author, *email, strings.Join(gitdirs, ", "))
	return nil
}

func identityRemove(env *env, args []string) error {
	name, _, err := takeArg(args, "name")
	if err != nil {
		return fmt.Errorf("%w\nusage: dotfiles-local identity rm <name>", err)
	}
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	if err := cfg.removeIdentity(name); err != nil {
		return err
	}
	if err := commit(env, cfg); err != nil {
		return err
	}
	fmt.Fprintf(env.out, "removed identity %s\n", name)
	return nil
}

func runHelper(env *env, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: dotfiles-local helper <set|rm> <host> ...")
	}
	switch args[0] {
	case "set":
		return helperSet(env, args[1:])
	case "rm":
		return helperRemove(env, args[1:])
	default:
		return fmt.Errorf("unknown helper command %q", args[0])
	}
}

func helperSet(env *env, args []string) error {
	if len(args) != 2 {
		return errors.New("usage: dotfiles-local helper set <host> <command>")
	}
	host, command := args[0], args[1]
	if host == "" || command == "" {
		return errors.New("host and command must both be non-empty")
	}
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg.setCredentialHelper(host, command)
	if err := commit(env, cfg); err != nil {
		return err
	}
	fmt.Fprintf(env.out, "credential helper for %s: %s\n", host, command)
	return nil
}

func helperRemove(env *env, args []string) error {
	host, _, err := takeArg(args, "host")
	if err != nil {
		return fmt.Errorf("%w\nusage: dotfiles-local helper rm <host>", err)
	}
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	if err := cfg.removeCredentialHelper(host); err != nil {
		return err
	}
	if err := commit(env, cfg); err != nil {
		return err
	}
	fmt.Fprintf(env.out, "removed credential helper for %s\n", host)
	return nil
}

func runProvider(env *env, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: dotfiles-local provider <add|rm> <alias> ...")
	}
	switch args[0] {
	case "add":
		return providerAdd(env, args[1:])
	case "rm":
		return providerRemove(env, args[1:])
	default:
		return fmt.Errorf("unknown provider command %q", args[0])
	}
}

func providerAdd(env *env, args []string) error {
	alias, rest, err := takeArg(args, "alias")
	if err != nil {
		return fmt.Errorf("%w\nusage: dotfiles-local provider add <alias> --type <github|gitlab> --host <host> [--namespace <ns>] [--default]", err)
	}
	fs := newFlagSet("provider add", env.err)
	kind := fs.String("type", "", `backend: "github" or "gitlab"`)
	host := fs.String("host", "", "hostname cln clones from")
	namespace := fs.String("namespace", "", "namespace used when a repository is given without one")
	makeDefault := fs.Bool("default", false, "also make this the default cln provider")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if *kind == "" || *host == "" {
		return errors.New("provider add requires --type and --host")
	}
	if *kind != "github" && *kind != "gitlab" {
		return fmt.Errorf("--type is %q, want \"github\" or \"gitlab\"", *kind)
	}

	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg.setProvider(alias, Provider{Type: *kind, Host: *host, DefaultNamespace: *namespace})
	if *makeDefault {
		cfg.setDefaultProvider(alias)
	}
	if err := commit(env, cfg); err != nil {
		return err
	}
	fmt.Fprintf(env.out, "provider %s: %s at %s\n", alias, *kind, *host)
	if *makeDefault {
		fmt.Fprintf(env.out, "default provider: %s\n", alias)
	}
	return nil
}

func providerRemove(env *env, args []string) error {
	alias, _, err := takeArg(args, "alias")
	if err != nil {
		return fmt.Errorf("%w\nusage: dotfiles-local provider rm <alias>", err)
	}
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	fellBack, err := cfg.removeProvider(alias)
	if err != nil {
		return err
	}
	if fellBack {
		// Leaving it would fail validation; the built-in gh provider is the
		// only alias guaranteed to exist.
		fmt.Fprintf(env.err, "default provider was %s; falling back to %s\n", alias, builtinProvider)
	}
	if err := commit(env, cfg); err != nil {
		return err
	}
	fmt.Fprintf(env.out, "removed provider %s\n", alias)
	return nil
}

func runDefaultProvider(env *env, args []string) error {
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		current := cfg.defaultProvider()
		if current == "" {
			current = builtinProvider + " (default)"
		}
		fmt.Fprintln(env.out, current)
		return nil
	}
	if len(args) > 1 {
		return errors.New("usage: dotfiles-local default-provider [<alias>]")
	}
	alias := args[0]
	cfg.setDefaultProvider(alias)
	if err := commit(env, cfg); err != nil {
		return err
	}
	fmt.Fprintf(env.out, "default provider: %s\n", alias)
	return nil
}

// commit validates before writing so a rejected change never reaches disk and
// never breaks the next `just switch`.
func commit(env *env, cfg Config) error {
	if err := cfg.validate(); err != nil {
		return fmt.Errorf("refusing to write an invalid configuration:\n%w", err)
	}
	return saveConfig(cfg)
}
