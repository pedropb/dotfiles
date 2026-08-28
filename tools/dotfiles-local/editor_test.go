package main

import (
	"strings"
	"testing"
)

// buildRows and configsEqual are the editor's only huh-free logic; the huh
// glue itself needs a real terminal and isn't unit tested, matching how
// `init`'s interactive body was never tested before it existed.

func TestBuildRowsListsEveryEditableFactAndTrailingActions(t *testing.T) {
	cfg := sampleConfig()
	rows := buildRows(cfg)

	var keys []string
	for _, row := range rows {
		keys = append(keys, row.key)
	}
	joined := strings.Join(keys, ",")

	for _, want := range []string{
		rowKeySystem, rowKeyUsername, rowKeyHomeDirectory,
		rowPrefixIdentity + "alt", rowPrefixIdentity + "work",
		rowKeyAddIdentity,
		rowPrefixHelper + "a.example.com",
		rowKeyAddHelper,
		rowPrefixProvider + "corp",
		rowKeyAddProvider,
		rowKeyDefaultProvider,
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("rows missing %q, got keys %v", want, keys)
		}
	}
	if keys[len(keys)-2] != rowKeySave || keys[len(keys)-1] != rowKeyQuit {
		t.Errorf("expected the last two rows to be save then quit, got %v", keys[len(keys)-2:])
	}
	// identities are sorted, so "alt" precedes "work".
	if strings.Index(joined, "alt") > strings.Index(joined, "work") {
		t.Errorf("expected identities sorted by name, got %v", keys)
	}
}

func TestBuildRowsOnEmptyConfigStillOffersEveryAddRow(t *testing.T) {
	rows := buildRows(Config{})
	var keys []string
	for _, row := range rows {
		keys = append(keys, row.key)
	}
	for _, want := range []string{rowKeyAddIdentity, rowKeyAddHelper, rowKeyAddProvider, rowKeyDefaultProvider} {
		found := false
		for _, k := range keys {
			if k == want {
				found = true
			}
		}
		if !found {
			t.Errorf("expected row %q on an empty config, got %v", want, keys)
		}
	}
}

func TestConfigsEqual(t *testing.T) {
	a := sampleConfig()
	b := sampleConfig()
	if !configsEqual(a, b) {
		t.Error("two configs built the same way should be equal")
	}
	b.Machine.Username = "someone-else"
	if configsEqual(a, b) {
		t.Error("a changed field should make configs unequal")
	}
}

func TestParseGitdirs(t *testing.T) {
	cases := map[string][]string{
		"~/src/a/":                {"~/src/a/"},
		"~/src/a/, ~/src/b/":      {"~/src/a/", "~/src/b/"},
		" ~/src/a/ ,, ~/src/b/ ,": {"~/src/a/", "~/src/b/"},
		"":                        nil,
		"   ":                     nil,
	}
	for input, want := range cases {
		got := parseGitdirs(input)
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("parseGitdirs(%q) = %v, want %v", input, got, want)
		}
	}
}

// keyCollision is the guard against a rename or an "+ Add" landing on a
// different entry and silently clobbering it — the mutation helpers
// (setIdentity etc.) have no such check themselves, since a plain edit that
// keeps its own key must go through unhindered.
func TestKeyCollision(t *testing.T) {
	cases := []struct {
		name                      string
		existingKey, newKey       string
		newKeyExists, wantCollide bool
	}{
		{"add with a free name", "", "new", false, false},
		{"add with a name already in use", "", "taken", true, true},
		{"edit without renaming", "work", "work", true, false},
		{"rename to a free name", "work", "personal", false, false},
		{"rename onto another entry", "work", "personal", true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := keyCollision(c.existingKey, c.newKey, c.newKeyExists)
			if got != c.wantCollide {
				t.Errorf("keyCollision(%q, %q, %v) = %v, want %v",
					c.existingKey, c.newKey, c.newKeyExists, got, c.wantCollide)
			}
		})
	}
}
