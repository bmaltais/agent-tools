package main

import (
	"bytes"
	"strings"
	"testing"
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
	code := run([]string{"open", "branch", "extra"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

func TestRun_EmptyBranch(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"open", ""}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "empty") {
		t.Fatalf("expected empty branch message, got %q", errOut.String())
	}
}

func TestRun_UnknownSubcommand(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"reopen", "mybranch"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "unknown subcommand") {
		t.Fatalf("expected unknown subcommand message, got %q", errOut.String())
	}
}

// TestRun_Open_GitNotInPath verifies that open exits 1 (not 2) when git is unavailable.
// This confirms arg parsing succeeded before the git call.
func TestRun_Open_GitNotInPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // empty PATH → git not found
	var out, errOut bytes.Buffer
	code := run([]string{"open", "feat/my-branch"}, &out, &errOut)
	// exit 1 = git not found; exit 2 = arg parse error (wrong)
	if code == 2 {
		t.Fatalf("arg parse should succeed (exit != 2), but got 2: %s", errOut.String())
	}
	if code != 1 {
		t.Fatalf("expected exit 1 (git not found), got %d", code)
	}
}

// TestRun_Close_GitNotInPath verifies that close exits 1 when git is unavailable.
func TestRun_Close_GitNotInPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // empty PATH → git not found
	var out, errOut bytes.Buffer
	code := run([]string{"close", "feat/my-branch"}, &out, &errOut)
	if code == 2 {
		t.Fatalf("arg parse should succeed (exit != 2), but got 2: %s", errOut.String())
	}
	if code != 1 {
		t.Fatalf("expected exit 1 (git not found), got %d", code)
	}
}
