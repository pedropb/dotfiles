package main

import "testing"

func TestFuzzyResolveExactMatch(t *testing.T) {
	got, err := fuzzyResolve("dotfiles", []string{"dotfiles", "other"})
	if err != nil || got != "dotfiles" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestFuzzyResolvePrefixMatch(t *testing.T) {
	got, err := fuzzyResolve("billing", []string{"billing-service", "reporting-tool"})
	if err != nil || got != "billing-service" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestFuzzyResolveNoCatalogIsLiteral(t *testing.T) {
	got, err := fuzzyResolve("anything", nil)
	if err != nil || got != "anything" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestFuzzyResolveNoMatchFallsBackToLiteral(t *testing.T) {
	got, err := fuzzyResolve("zzz", []string{"aaa", "bbb"})
	if err != nil || got != "zzz" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestFuzzyResolveAmbiguous(t *testing.T) {
	_, err := fuzzyResolve("billing", []string{"billing-service", "billing-web"})
	if err == nil {
		t.Fatal("expected an ambiguous-repository error")
	}
}

func TestIsSubsequence(t *testing.T) {
	cases := []struct {
		query, candidate string
		want             bool
	}{
		{"bsv", "billing-service", true},
		{"service", "billing-service", true},
		{"xyz", "billing-service", false},
		{"", "anything", true},
	}
	for _, c := range cases {
		if got := isSubsequence(c.query, c.candidate); got != c.want {
			t.Errorf("isSubsequence(%q, %q) = %v, want %v", c.query, c.candidate, got, c.want)
		}
	}
}
