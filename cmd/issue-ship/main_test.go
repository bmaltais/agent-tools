package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"implement issue-ship", "implement-issueship"},
		{"implement gh-merge-wait", "implement-ghmergewait"},
		{"Add New Feature With A Very Long Name", "add-new-feature-with-a"},
		{"  leading and trailing spaces  ", "leading-and-trailing-spaces"},
		{"special!@# chars", "special-chars"},
		{"one", "one"},
	}
	for _, c := range cases {
		got := slugify(c.input)
		if got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestBranchName(t *testing.T) {
	got := branchName(5, "implement issue-ship")
	want := "feat/issue-5-implement-issueship"
	if got != want {
		t.Errorf("branchName(5, ...) = %q, want %q", got, want)
	}
}

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

func TestRun_InvalidIssueNumber(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"owner/repo", "abc"}, &out, &errOut)
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
		t.Fatalf("expected method error, got %q", errOut.String())
	}
}

func TestRun_InvalidFromStage(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--from-stage", "bogus", "owner/repo", "1"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "--from-stage") {
		t.Fatalf("expected from-stage error, got %q", errOut.String())
	}
}

func TestValidStage(t *testing.T) {
	for _, s := range []string{"branch", "pr", "merge", "cleanup"} {
		if !validStage(s) {
			t.Errorf("validStage(%q) = false, want true", s)
		}
	}
	if validStage("bogus") {
		t.Error("validStage(\"bogus\") = true, want false")
	}
}

func TestStageIndex(t *testing.T) {
	cases := []struct {
		stage string
		want  int
	}{
		{"branch", 0},
		{"pr", 1},
		{"merge", 2},
		{"cleanup", 3},
	}
	for _, c := range cases {
		got := stageIndex(c.stage)
		if got != c.want {
			t.Errorf("stageIndex(%q) = %d, want %d", c.stage, got, c.want)
		}
	}
}

func TestLogStage_DryRun(t *testing.T) {
	var out bytes.Buffer
	cfg := &config{
		dryRun: true,
		stdout: &out,
		stderr: &out,
	}
	cfg.logStage("branch", "create feat/issue-5-foo from main")
	got := out.String()
	if !strings.Contains(got, "[dry-run]") {
		t.Errorf("expected [dry-run] prefix, got %q", got)
	}
	if !strings.Contains(got, "[branch]") {
		t.Errorf("expected [branch] tag, got %q", got)
	}
}

func TestLogStage_NoDryRun(t *testing.T) {
	var out bytes.Buffer
	cfg := &config{
		dryRun: false,
		stdout: &out,
		stderr: &out,
	}
	cfg.logStage("pr", "open PR for feat/issue-5-foo")
	got := out.String()
	if strings.Contains(got, "[dry-run]") {
		t.Errorf("unexpected [dry-run] prefix in non-dry-run mode: %q", got)
	}
	if !strings.Contains(got, "[pr]") {
		t.Errorf("expected [pr] tag, got %q", got)
	}
}
