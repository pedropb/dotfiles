// cln clones git repositories using a short provider/namespace/repo
// shorthand, e.g. `cln gh dotfiles` or `cln team/billing`, instead of
// typing the full HTTPS remote URL. See README.md.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const usage = `usage: cln [-n] [provider] <[namespace/]repo> [dest]
       cln providers

  -n, --dry-run   print the resolved clone URL instead of cloning

Without an explicit dest, clones land in
  ~/src/<host>/<namespace>/<repo>
the layout config/zsh/scd.zsh expects, so cloned repos are immediately
"cd"-able by "scd <namespace>/<repo>".

Examples:
  cln gh dotfiles          clone into ~/src/github.com/<gh's default namespace>/dotfiles
  cln octocat/Hello-World  clone into ~/src/github.com/octocat/Hello-World
  cln team/billing         clone the default provider's team/billing,
                           fuzzy-matching "billing" against team's repos
  cln providers            list configured providers
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "cln: "+err.Error())
		os.Exit(1)
	}
}

func run(argv []string) error {
	fs := flag.NewFlagSet("cln", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	var dryRun bool
	fs.BoolVar(&dryRun, "n", false, "print the resolved clone URL instead of cloning")
	fs.BoolVar(&dryRun, "dry-run", false, "print the resolved clone URL instead of cloning")

	if err := fs.Parse(argv); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	args := fs.Args()

	if len(args) == 0 {
		fs.Usage()
		return errors.New("no repository given")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if len(args) == 1 && args[0] == "providers" {
		printProviders(cfg)
		return nil
	}

	sp, err := parseArgs(args, cfg.Providers)
	if err != nil {
		return err
	}

	res, err := resolve(cfg, sp)
	if err != nil {
		return err
	}
	if res.repo != res.query {
		fmt.Fprintf(os.Stderr, "cln: matched %q to %q\n", res.query, res.repo)
	}

	dest := sp.dest
	if dest == "" {
		dest = defaultDest(res.provider.Host, res.namespace, res.repo)
	}
	cloneArgs := []string{"clone", cloneURL(res.provider, res.namespace, res.repo), dest}

	if dryRun {
		fmt.Println("git " + strings.Join(cloneArgs, " "))
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(dest), err)
	}
	return gitClone(cloneArgs)
}

// gitClone execs git, replacing cln's own exit code with git's on failure so
// git's own error message stands alone instead of being wrapped again.
func gitClone(args []string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func printProviders(cfg Config) {
	aliases := make([]string, 0, len(cfg.Providers))
	for alias := range cfg.Providers {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)

	for _, alias := range aliases {
		p := cfg.Providers[alias]
		namespace := p.DefaultNamespace
		if namespace == "" {
			namespace = "-"
		}
		marker := ""
		if alias == cfg.DefaultProvider {
			marker = "  (default)"
		}
		fmt.Printf("%-10s %-8s %-24s default-namespace=%s%s\n", alias, p.Type, p.Host, namespace, marker)
	}
}
