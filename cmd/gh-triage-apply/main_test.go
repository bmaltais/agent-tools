package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestRun_MissingPositional verifies usage is printed when too few args are given.
func TestRun_MissingPositional(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--labels", "bug"}, nil, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "usage:") {
		t.Fatalf("expected usage message, got %q", errOut.String())
	}
}

// TestRun_InvalidIssueNumber verifies non-integer issue numbers are rejected.
func TestRun_InvalidIssueNumber(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--labels", "bug", "--comment", "hi", "owner/repo", "abc"}, nil, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "invalid issue number") {
		t.Fatalf("expected invalid issue number message, got %q", errOut.String())
	}
}

// TestRun_ZeroIssueNumber verifies issue number 0 is rejected.
func TestRun_ZeroIssueNumber(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--labels", "bug", "--comment", "hi", "owner/repo", "0"}, nil, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

// TestRun_MissingLabels verifies --labels is required.
func TestRun_MissingLabels(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--comment", "hello", "owner/repo", "1"}, nil, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "--labels") {
		t.Fatalf("expected --labels error, got %q", errOut.String())
	}
}

// TestRun_MissingComment verifies that one of --comment or --comment-file is required.
func TestRun_MissingComment(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--labels", "bug", "owner/repo", "1"}, nil, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "--comment") {
		t.Fatalf("expected comment error, got %q", errOut.String())
	}
}

// TestRun_MutuallyExclusiveComment verifies --comment and --comment-file cannot both be set.
func TestRun_MutuallyExclusiveComment(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--labels", "bug", "--comment", "hi", "--comment-file", "foo.txt", "owner/repo", "1"}, nil, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "mutually exclusive") {
		t.Fatalf("expected mutually exclusive message, got %q", errOut.String())
	}
}

// TestRun_DryRun_InlineComment verifies dry-run prints both operations without calling gh.
func TestRun_DryRun_InlineComment(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{
		"--labels", "enhancement,ready-for-agent",
		"--comment", "triage note",
		"--dry-run",
		"owner/repo", "42",
	}, nil, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, errOut.String())
	}
	got := out.String()
	if !strings.Contains(got, "[dry-run]") {
		t.Errorf("expected [dry-run] prefix, got %q", got)
	}
	if !strings.Contains(got, "issue edit") {
		t.Errorf("expected issue edit in dry-run output, got %q", got)
	}
	if !strings.Contains(got, "issue comment") {
		t.Errorf("expected issue comment in dry-run output, got %q", got)
	}
	if !strings.Contains(got, "enhancement,ready-for-agent") {
		t.Errorf("expected label in dry-run output, got %q", got)
	}
	if !strings.Contains(got, "triage note") {
		t.Errorf("expected comment body in dry-run output, got %q", got)
	}
}

// TestRun_DryRun_CommentFileStdin verifies --comment-file - reads from stdin in dry-run.
func TestRun_DryRun_CommentFileStdin(t *testing.T) {
	stdinBody := "body from stdin"
	var out, errOut bytes.Buffer
	code := run([]string{
		"--labels", "bug",
		"--comment-file", "-",
		"--dry-run",
		"owner/repo", "7",
	}, strings.NewReader(stdinBody), &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), stdinBody) {
		t.Errorf("expected stdin body in dry-run output, got %q", out.String())
	}
}

// TestReadCommentBody_Inline verifies inline text is returned directly.
func TestReadCommentBody_Inline(t *testing.T) {
	got, err := readCommentBody("hello", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

// TestReadCommentBody_Stdin verifies - reads from the provided reader.
func TestReadCommentBody_Stdin(t *testing.T) {
	r := strings.NewReader("from stdin")
	got, err := readCommentBody("", "-", r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "from stdin" {
		t.Errorf("got %q, want %q", got, "from stdin")
	}
}

// TestReadCommentBody_File verifies a real file path is read correctly.
func TestReadCommentBody_File(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/comment.txt"
	content := "comment from file"
	if err := writeFile(path, content); err != nil {
		t.Fatalf("setup: %v", err)
	}
	got, err := readCommentBody("", path, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

// TestReadCommentBody_FileMissing verifies a missing file returns an error.
func TestReadCommentBody_FileMissing(t *testing.T) {
	_, err := readCommentBody("", "/nonexistent/path/comment.txt", nil)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// TestRun_DryRun_EmptyCommentBody verifies an empty body after --comment-file is rejected.
func TestRun_DryRun_EmptyCommentBody(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{
		"--labels", "bug",
		"--comment-file", "-",
		"--dry-run",
		"owner/repo", "1",
	}, strings.NewReader("   "), &out, &errOut)
	if code == 0 {
		t.Fatal("expected non-zero exit for empty comment body, got 0")
	}
}

// writeFile is a test helper that writes content to a file.
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o600)
}
