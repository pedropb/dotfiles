package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// maxCatalogSize caps how many repository names cln will hold for a single
// namespace. A namespace at or beyond the cap is treated as unavailable —
// fuzzyResolve then falls back to the literal repository name — rather
// than risking a partial listing silently steering a fuzzy match.
const maxCatalogSize = 300

// fetchCatalog lists repository slugs in namespace by shelling out to the
// provider's already-authenticated CLI, reusing whatever credentials that
// CLI has stored — see README.md's "Authentication" section. Adding a
// backend means adding a case here, a fetchXCatalog following the same
// shape, and a Config.validate case in config.go.
func fetchCatalog(p Provider, namespace string) ([]string, error) {
	switch p.Type {
	case "github":
		return fetchGitHubCatalog(p, namespace)
	case "gitlab":
		return fetchGitLabCatalog(p, namespace)
	default:
		return nil, fmt.Errorf("provider %q: unknown type %q", p.Alias, p.Type)
	}
}

// fetchGitHubCatalog lists repository names owned by namespace (a user or
// organization) using the gh CLI. GH_HOST targets a GitHub Enterprise host
// when p.Host isn't github.com.
func fetchGitHubCatalog(p Provider, namespace string) ([]string, error) {
	cmd := exec.Command("gh", "repo", "list", namespace,
		"--json", "name", "--jq", ".[].name",
		"-L", strconv.Itoa(maxCatalogSize+1))
	cmd.Env = append(os.Environ(), "GH_HOST="+p.Host)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	names := splitLines(out)
	if len(names) > maxCatalogSize {
		return nil, fmt.Errorf("namespace %q has too many repositories to list", namespace)
	}
	return names, nil
}

// fetchGitLabCatalog lists project paths directly in namespace (a group,
// which may itself be nested, e.g. "team/sub") using the glab CLI.
func fetchGitLabCatalog(p Provider, namespace string) ([]string, error) {
	endpoint := fmt.Sprintf("groups/%s/projects?per_page=100&simple=true", url.PathEscape(namespace))
	cmd := exec.Command("glab", "api", endpoint,
		"--hostname", p.Host, "--paginate", "--output", "ndjson")

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, line := range splitLines(out) {
		var project struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(line), &project); err != nil {
			return nil, fmt.Errorf("parsing glab output: %w", err)
		}
		names = append(names, project.Path)
	}
	if len(names) > maxCatalogSize {
		return nil, fmt.Errorf("namespace %q has too many repositories to list", namespace)
	}
	return names, nil
}

func splitLines(b []byte) []string {
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}
