package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// prompter is a line-oriented question/answer helper. A full-screen TUI would
// need a dependency and would not survive a non-interactive shell; `init` is
// only ever a convenience wrapper around the flag-driven commands, which stay
// the scriptable path.
type prompter struct {
	in  *bufio.Reader
	out io.Writer
}

func (p *prompter) line(question, fallback string) (string, error) {
	if fallback == "" {
		fmt.Fprintf(p.out, "%s: ", question)
	} else {
		fmt.Fprintf(p.out, "%s [%s]: ", question, fallback)
	}
	text, err := p.in.ReadString('\n')
	if err != nil && (err != io.EOF || text == "") {
		return "", errors.New("input ended before the answer")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return fallback, nil
	}
	return text, nil
}

func (p *prompter) required(question string) (string, error) {
	for {
		answer, err := p.line(question, "")
		if err != nil {
			return "", err
		}
		if answer != "" {
			return answer, nil
		}
		fmt.Fprintln(p.out, "  a value is required")
	}
}

func (p *prompter) confirm(question string, fallback bool) (bool, error) {
	hint := "y/N"
	if fallback {
		hint = "Y/n"
	}
	for {
		fmt.Fprintf(p.out, "%s [%s]: ", question, hint)
		text, err := p.in.ReadString('\n')
		if err != nil && (err != io.EOF || text == "") {
			return false, errors.New("input ended before the answer")
		}
		switch strings.ToLower(strings.TrimSpace(text)) {
		case "":
			return fallback, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		}
		fmt.Fprintln(p.out, "  answer y or n")
	}
}

func (p *prompter) choice(question string, options []string, fallback string) (string, error) {
	for {
		answer, err := p.line(fmt.Sprintf("%s (%s)", question, strings.Join(options, "/")), fallback)
		if err != nil {
			return "", err
		}
		for _, option := range options {
			if answer == option {
				return answer, nil
			}
		}
		fmt.Fprintf(p.out, "  choose one of: %s\n", strings.Join(options, ", "))
	}
}

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

	p := &prompter{in: bufio.NewReader(env.in), out: env.out}

	fmt.Fprintf(env.out, "Configuring %s\n\n", path)
	fmt.Fprintf(env.out, "This machine: %s, user %s, home %s\n",
		cfg.Machine.System, cfg.Machine.Username, cfg.Machine.HomeDirectory)
	keep, err := p.confirm("Use these", true)
	if err != nil {
		return err
	}
	if !keep {
		if cfg.Machine.System, err = p.choice("system", knownSystems, cfg.Machine.System); err != nil {
			return err
		}
		if cfg.Machine.Username, err = p.line("username", cfg.Machine.Username); err != nil {
			return err
		}
		if cfg.Machine.HomeDirectory, err = p.line("home directory", cfg.Machine.HomeDirectory); err != nil {
			return err
		}
	}

	if err := promptIdentities(p, env, &cfg); err != nil {
		return err
	}
	if err := promptHelpers(p, env, &cfg); err != nil {
		return err
	}
	if err := promptProviders(p, env, &cfg); err != nil {
		return err
	}

	preview, err := encodeConfig(cfg)
	if err != nil {
		return err
	}
	fmt.Fprintf(env.out, "\n%s\n", preview)
	verb := "Write"
	if existed {
		verb = "Replace"
	}
	write, err := p.confirm(fmt.Sprintf("%s %s", verb, path), true)
	if err != nil {
		return err
	}
	if !write {
		fmt.Fprintln(env.out, "nothing written")
		return nil
	}
	if err := commit(env, cfg); err != nil {
		return err
	}
	fmt.Fprintf(env.out, "wrote %s\nrun `just switch` to apply\n", path)
	return nil
}

func promptIdentities(p *prompter, env *env, cfg *Config) error {
	fmt.Fprintln(env.out, "\nGit identities — author name and email applied under a directory prefix.")
	for _, name := range sortedKeys(cfg.identities()) {
		identity := cfg.identities()[name]
		fmt.Fprintf(env.out, "  have %s: %s <%s> under %s\n", name, identity.Name, identity.Email, strings.Join(identity.Gitdir, ", "))
	}
	for {
		add, err := p.confirm("Add a git identity", len(cfg.identities()) == 0)
		if err != nil || !add {
			return err
		}
		name, err := p.required("  short name (e.g. work)")
		if err != nil {
			return err
		}
		author, err := p.required("  author name")
		if err != nil {
			return err
		}
		email, err := p.required("  author email")
		if err != nil {
			return err
		}
		raw, err := p.required("  directory prefixes, comma separated (e.g. ~/src/git.example.com/)")
		if err != nil {
			return err
		}
		var gitdirs []string
		for _, part := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				gitdirs = append(gitdirs, trimmed)
			}
		}
		git := cfg.gitSection()
		if git.Identities == nil {
			git.Identities = map[string]Identity{}
		}
		git.Identities[name] = Identity{Gitdir: gitdirs, Name: author, Email: email}
	}
}

func promptHelpers(p *prompter, env *env, cfg *Config) error {
	fmt.Fprintln(env.out, "\nCredential helpers — the git credential command for one HTTPS host.")
	fmt.Fprintln(env.out, "  github.com is already wired to `gh auth git-credential`.")
	for _, host := range sortedKeys(cfg.credentialHelpers()) {
		fmt.Fprintf(env.out, "  have %s: %s\n", host, cfg.credentialHelpers()[host])
	}
	for {
		add, err := p.confirm("Add a credential helper", false)
		if err != nil || !add {
			return err
		}
		host, err := p.required("  host (exactly as the remote URL spells it)")
		if err != nil {
			return err
		}
		command, err := p.line("  helper command", "!glab auth git-credential")
		if err != nil {
			return err
		}
		git := cfg.gitSection()
		if git.CredentialHelpers == nil {
			git.CredentialHelpers = map[string]string{}
		}
		git.CredentialHelpers[host] = command
	}
}

func promptProviders(p *prompter, env *env, cfg *Config) error {
	fmt.Fprintln(env.out, "\ncln providers — forges `cln` can clone from by short alias.")
	fmt.Fprintf(env.out, "  %s (github.com) is built in.\n", builtinProvider)
	for _, alias := range sortedKeys(cfg.providers()) {
		provider := cfg.providers()[alias]
		fmt.Fprintf(env.out, "  have %s: %s at %s\n", alias, provider.Type, provider.Host)
	}
	for {
		add, err := p.confirm("Add a cln provider", false)
		if err != nil || !add {
			break
		}
		alias, err := p.required("  alias (e.g. corp)")
		if err != nil {
			return err
		}
		kind, err := p.choice("  type", []string{"github", "gitlab"}, "gitlab")
		if err != nil {
			return err
		}
		host, err := p.required("  host")
		if err != nil {
			return err
		}
		namespace, err := p.line("  default namespace (optional)", "")
		if err != nil {
			return err
		}
		cln := cfg.clnSection()
		if cln.Providers == nil {
			cln.Providers = map[string]Provider{}
		}
		cln.Providers[alias] = Provider{Type: kind, Host: host, DefaultNamespace: namespace}
	}

	if len(cfg.providers()) == 0 {
		return nil
	}
	options := append([]string{builtinProvider}, sortedKeys(cfg.providers())...)
	current := cfg.defaultProvider()
	if current == "" {
		current = builtinProvider
	}
	chosen, err := p.choice("Default provider for a bare `cln <namespace>/<repo>`", options, current)
	if err != nil {
		return err
	}
	if chosen == builtinProvider {
		cfg.clnSection().DefaultProvider = ""
	} else {
		cfg.clnSection().DefaultProvider = chosen
	}
	return nil
}

func interactive() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
