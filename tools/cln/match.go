package main

import (
	"fmt"
	"sort"
	"strings"
)

// fuzzyResolve picks the catalog entry that best matches query. An empty
// catalog (no listing available, or too large to search — see
// provider.go's maxCatalogSize) disables fuzzy matching entirely: query is
// returned unchanged, and git itself will report an unknown repository if
// it turns out not to exist.
func fuzzyResolve(query string, catalog []string) (string, error) {
	if len(catalog) == 0 {
		return query, nil
	}

	best, ambiguous := pickTier(query, catalog)
	if ambiguous != nil {
		return "", fmt.Errorf("ambiguous repository %q: %s", query, strings.Join(ambiguous, ", "))
	}
	if best == "" {
		return query, nil
	}
	return best, nil
}

// pickTier ranks candidates by match quality — exact, prefix, substring,
// then subsequence, the same family of match fzf uses — and returns the
// sole winner from the best non-empty tier. Multiple winners in that tier
// are reported as ambiguous rather than guessed at, mirroring how
// config/zsh/scd.zsh handles ambiguous directory matches.
func pickTier(query string, candidates []string) (best string, ambiguous []string) {
	q := strings.ToLower(query)
	var exact, prefix, substring, subsequence []string

	for _, c := range candidates {
		lc := strings.ToLower(c)
		switch {
		case lc == q:
			exact = append(exact, c)
		case strings.HasPrefix(lc, q):
			prefix = append(prefix, c)
		case strings.Contains(lc, q):
			substring = append(substring, c)
		case isSubsequence(q, lc):
			subsequence = append(subsequence, c)
		}
	}

	for _, tier := range [][]string{exact, prefix, substring, subsequence} {
		switch len(tier) {
		case 0:
			continue
		case 1:
			return tier[0], nil
		default:
			sort.Strings(tier)
			return "", tier
		}
	}
	return "", nil
}

// isSubsequence reports whether every character of query appears in
// candidate in order, not necessarily contiguously. Both arguments are
// expected to already be lowercased.
func isSubsequence(query, candidate string) bool {
	i := 0
	for j := 0; j < len(candidate) && i < len(query); j++ {
		if candidate[j] == query[i] {
			i++
		}
	}
	return i == len(query)
}
