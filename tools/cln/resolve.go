package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// spec is the user's request before provider and namespace defaults are
// applied.
type spec struct {
	providerAlias string // "" selects cfg.DefaultProvider
	repoPath      string // "[namespace/]repo"
	dest          string // "" lets git pick the directory name
}

// parseArgs disambiguates cln's positional arguments. A 2-argument
// invocation is a provider+repo pair only when the first argument names a
// configured provider; otherwise it is a repo+dest pair, matching how
// `git clone <repo> [dir]` already behaves.
func parseArgs(args []string, providers map[string]Provider) (spec, error) {
	switch len(args) {
	case 1:
		return spec{repoPath: args[0]}, nil
	case 2:
		if _, ok := providers[args[0]]; ok {
			return spec{providerAlias: args[0], repoPath: args[1]}, nil
		}
		return spec{repoPath: args[0], dest: args[1]}, nil
	case 3:
		if _, ok := providers[args[0]]; !ok {
			return spec{}, fmt.Errorf("unknown provider %q", args[0])
		}
		return spec{providerAlias: args[0], repoPath: args[1], dest: args[2]}, nil
	default:
		return spec{}, fmt.Errorf("usage: cln [provider] <[namespace/]repo> [dest]")
	}
}

// resolved is a spec with provider and namespace defaults applied and, when
// a catalog was available, fuzzy matching performed.
type resolved struct {
	provider  Provider
	namespace string
	repo      string
	query     string // repo as requested, before fuzzy matching
}

func resolve(cfg Config, sp spec) (resolved, error) {
	alias := sp.providerAlias
	if alias == "" {
		alias = cfg.DefaultProvider
	}
	provider, ok := cfg.Providers[alias]
	if !ok {
		return resolved{}, fmt.Errorf("unknown provider %q", alias)
	}

	namespace, query, err := splitRepoPath(sp.repoPath, provider)
	if err != nil {
		return resolved{}, err
	}

	catalog := catalogFor(provider, namespace)
	repo, err := fuzzyResolve(query, catalog)
	if err != nil {
		return resolved{}, err
	}

	return resolved{provider: provider, namespace: namespace, repo: repo, query: query}, nil
}

func splitRepoPath(repoPath string, p Provider) (namespace, repo string, err error) {
	idx := strings.LastIndex(repoPath, "/")
	if idx == -1 {
		if p.DefaultNamespace == "" {
			return "", "", fmt.Errorf("provider %q has no default namespace; use '<namespace>/%s'", p.Alias, repoPath)
		}
		return p.DefaultNamespace, repoPath, nil
	}

	namespace, repo = repoPath[:idx], repoPath[idx+1:]
	if namespace == "" || repo == "" {
		return "", "", fmt.Errorf("invalid repository %q", repoPath)
	}
	if p.Type == "github" && strings.Contains(namespace, "/") {
		return "", "", fmt.Errorf("provider %q (github) does not support nested namespaces: %q", p.Alias, namespace)
	}
	return namespace, repo, nil
}

func cloneURL(p Provider, namespace, repo string) string {
	return fmt.Sprintf("https://%s/%s/%s", p.Host, namespace, repo)
}

// defaultDest is where a clone lands when the caller doesn't give an
// explicit destination: ~/src/<host>/<namespace>/<repo>, the same layout
// config/zsh/scd.zsh expects under $HOME/src (three path components:
// host, then namespace, then repo). An explicit destination argument
// always overrides this.
func defaultDest(host, namespace, repo string) string {
	return filepath.Join(os.Getenv("HOME"), "src", host, namespace, repo)
}
