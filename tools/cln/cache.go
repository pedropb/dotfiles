package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// cacheTTL bounds how long a namespace's catalog is trusted before cln
// refetches it. A refetch failure (offline, revoked auth, ...) falls back
// to whatever is on disk, however stale, rather than disabling fuzzy
// matching outright.
const cacheTTL = time.Hour

type cacheEntry struct {
	FetchedAt time.Time `json:"fetched_at"`
	Repos     []string  `json:"repos"`
}

func cacheDir() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "cln")
	}
	return filepath.Join(os.Getenv("HOME"), ".cache", "cln")
}

func cachePath(alias, namespace string) string {
	safe := strings.ReplaceAll(namespace, "/", "_")
	return filepath.Join(cacheDir(), alias+"_"+safe+".json")
}

func readCacheEntry(path string) (cacheEntry, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return cacheEntry{}, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return cacheEntry{}, false
	}
	return entry, true
}

func writeCacheEntry(path string, repos []string) {
	data, err := json.Marshal(cacheEntry{FetchedAt: time.Now(), Repos: repos})
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

// catalogFor returns namespace's repository listing, maintaining it on disk
// across invocations: it refreshes the cache when missing or older than
// cacheTTL, and otherwise serves the cached copy directly. A nil result
// means no catalog is available at all (never fetched successfully, and
// nothing cached); callers must treat that as "skip fuzzy matching", not
// as an error.
func catalogFor(p Provider, namespace string) []string {
	path := cachePath(p.Alias, namespace)

	if entry, ok := readCacheEntry(path); ok && time.Since(entry.FetchedAt) < cacheTTL {
		return entry.Repos
	}

	if repos, err := fetchCatalog(p, namespace); err == nil {
		writeCacheEntry(path, repos)
		return repos
	}

	if entry, ok := readCacheEntry(path); ok {
		return entry.Repos
	}
	return nil
}
