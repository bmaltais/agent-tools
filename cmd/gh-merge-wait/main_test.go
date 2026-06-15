package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestRun_MissingArgs(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "usage:") {
		t.Fatalf("expected usage message, got %q", errOut.String())
	}
}

func TestRun_TooManyArgs(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"owner/repo", "1", "extra"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

func TestRun_InvalidPRNumber(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"owner/repo", "abc"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "invalid PR number") {
		t.Fatalf("expected invalid PR number message, got %q", errOut.String())
	}
}

func TestRun_ZeroPRNumber(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"owner/repo", "0"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

func TestRun_InvalidMethod(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--method", "fast-forward", "owner/repo", "1"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "--method") {
		t.Fatalf("expected --method error, got %q", errOut.String())
	}
}

func TestRun_InvalidTimeout(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--timeout", "notaduration", "owner/repo", "1"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

func TestIs502Output(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"HTTP 502 Bad Gateway", true},
		{"error: HTTP 502", true},
		{"502", true},
		{"HTTP 503 Service Unavailable", false},
		{"", false},
		{"Request failed with status 200", false},
	}
	for _, c := range cases {
		got := is502Output(c.input)
		if got != c.want {
			t.Errorf("is502Output(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestParsePullResponse_Merged(t *testing.T) {
	data := map[string]interface{}{
		"merged":           true,
		"merge_commit_sha": "abc123def456",
	}
	b, _ := json.Marshal(data)

	var pr pullResponse
	if err := json.Unmarshal(b, &pr); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !pr.Merged {
		t.Error("expected Merged=true")
	}
	if pr.MergeCommitSHA != "abc123def456" {
		t.Errorf("expected SHA abc123def456, got %q", pr.MergeCommitSHA)
	}
}

func TestParsePullResponse_NotYetMerged(t *testing.T) {
	data := map[string]interface{}{
		"merged":           false,
		"merge_commit_sha": nil,
	}
	b, _ := json.Marshal(data)

	var pr pullResponse
	if err := json.Unmarshal(b, &pr); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if pr.Merged {
		t.Error("expected Merged=false")
	}
	if pr.MergeCommitSHA != "" {
		t.Errorf("expected empty SHA, got %q", pr.MergeCommitSHA)
	}
}

func TestPollUntilMerged_TimeoutImmediate(t *testing.T) {
	// deadline already in the past → should return timeout error immediately
	// without making any API calls (since gh binary is not available in tests)
	past := time.Now().Add(-1 * time.Second)
	_, err := pollUntilMerged("gh-does-not-exist-in-test", "owner/repo", 1, past)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout in error, got %q", err.Error())
	}
}

func TestRun_FlagsBeforePositionals(t *testing.T) {
	// flags must be parseable before positionals (Go flag package convention)
	// This test verifies --method and --timeout flags are accepted when placed before positionals.
	// We use a very short timeout so the gh binary lookup fails fast (no gh in test env is fine —
	// it will exit 1 from safeexec, which is the expected code path for missing binary).
	var out, errOut bytes.Buffer
	// Using a non-existent binary path via a short timeout won't work here —
	// the test just validates that the flags themselves parse without error (code != 2).
	// The actual merge attempt will fail with exit 1 (no gh binary), not 2.
	code := run([]string{"--method", "rebase", "--timeout", "1ms", "owner/repo", "42"}, &out, &errOut)
	// exit 1 means flags parsed OK but gh was not found or merge failed — that's acceptable
	// exit 2 means flag parse error — that would be a bug
	if code == 2 {
		t.Fatalf("flag parse failed (exit 2): %s", errOut.String())
	}
}
