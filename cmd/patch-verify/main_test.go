package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_ReplacesFirstOccurrenceAndWritesFile(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "input.txt")
	if err := os.WriteFile(file, []byte("alpha beta\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{file, "beta", "gamma"}, strings.NewReader(""), &out, &errOut)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr=%q", code, errOut.String())
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(got) != "alpha gamma\n" {
		t.Fatalf("file mismatch: got %q", string(got))
	}

	diff := out.String()
	if !strings.Contains(diff, "--- "+file+" (before)") || !strings.Contains(diff, "+++ "+file+" (after)") {
		t.Fatalf("expected unified diff headers, got %q", diff)
	}
	if !strings.Contains(diff, "-alpha beta") || !strings.Contains(diff, "+alpha gamma") {
		t.Fatalf("expected changed lines in diff, got %q", diff)
	}
}

func TestRun_NotFoundReturnsError(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "input.txt")
	if err := os.WriteFile(file, []byte("alpha beta\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{file, "missing", "gamma"}, strings.NewReader(""), &out, &errOut)
	if code == 0 {
		t.Fatalf("expected non-zero exit code")
	}
	if !strings.Contains(errOut.String(), "pattern not found") {
		t.Fatalf("expected not-found error, got %q", errOut.String())
	}
}

func TestRun_AmbiguousWithoutAllReturnsError(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "input.txt")
	if err := os.WriteFile(file, []byte("beta beta\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{file, "beta", "gamma"}, strings.NewReader(""), &out, &errOut)
	if code == 0 {
		t.Fatalf("expected non-zero exit code")
	}
	if !strings.Contains(errOut.String(), "use --all") {
		t.Fatalf("expected ambiguous-match guidance, got %q", errOut.String())
	}
}

func TestRun_AllReplacesEveryOccurrence(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "input.txt")
	if err := os.WriteFile(file, []byte("beta beta beta\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"--all", file, "beta", "gamma"}, strings.NewReader(""), &out, &errOut)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr=%q", code, errOut.String())
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(got) != "gamma gamma gamma\n" {
		t.Fatalf("file mismatch: got %q", string(got))
	}
}

func TestRun_DryRunDoesNotModifyFile(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "input.txt")
	before := "alpha beta\n"
	if err := os.WriteFile(file, []byte(before), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"--dry-run", file, "beta", "gamma"}, strings.NewReader(""), &out, &errOut)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr=%q", code, errOut.String())
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(got) != before {
		t.Fatalf("dry-run changed file: got %q", string(got))
	}
	if !strings.Contains(out.String(), "+alpha gamma") {
		t.Fatalf("expected diff output in dry-run, got %q", out.String())
	}
}

func TestRun_PreservesFileMode(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "input.sh")
	if err := os.WriteFile(file, []byte("echo beta\n"), 0o755); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{file, "beta", "gamma"}, strings.NewReader(""), &out, &errOut)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr=%q", code, errOut.String())
	}

	fi, err := os.Stat(file)
	if err != nil {
		t.Fatalf("stat output file: %v", err)
	}
	if fi.Mode().Perm() != 0o755 {
		t.Fatalf("file mode changed: got %o, want %o", fi.Mode().Perm(), os.FileMode(0o755))
	}
}

func TestRun_StdinModeWritesDiffAndPatchedContent(t *testing.T) {
	in := "alpha beta\n"

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"-", "beta", "gamma"}, strings.NewReader(in), &out, &errOut)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr=%q", code, errOut.String())
	}

	got := out.String()
	if !strings.Contains(got, "--- stdin (before)") || !strings.Contains(got, "+++ stdin (after)") {
		t.Fatalf("missing stdin diff headers: %q", got)
	}
	if !strings.Contains(got, "alpha gamma") {
		t.Fatalf("expected patched content on stdout, got %q", got)
	}
}

func TestRun_PipedOutputIsPlain(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "input.txt")
	if err := os.WriteFile(file, []byte("alpha beta\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{file, "beta", "gamma"}, strings.NewReader(""), &out, &errOut)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr=%q", code, errOut.String())
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("expected plain diff output when piped, got %q", out.String())
	}
}

func TestColorizeDiffAddsAnsiSequences(t *testing.T) {
	diff := "@@ -1 +1 @@\n-old\n+new\n"
	colored := colorizeDiff(diff)
	if !strings.Contains(colored, "\x1b[31m") || !strings.Contains(colored, "\x1b[32m") || !strings.Contains(colored, "\x1b[36m") {
		t.Fatalf("expected ANSI color sequences, got %q", colored)
	}
}
