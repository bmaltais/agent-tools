// gh-triage-apply applies a label set and posts a comment to one or more GitHub
// issues in a single command, replacing the two-call sequence of gh issue edit +
// gh issue comment. Multiple issue numbers are processed sequentially; all errors
// are collected and reported — execution is not stopped on the first failure.
//
// Usage: gh-triage-apply [--dry-run] --labels <label,...> (--comment <text> | --comment-file <path|->) <owner/repo> <issue-number> [<issue-number>...]
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cli/safeexec"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gh-triage-apply", flag.ContinueOnError)
	fs.SetOutput(stderr)
	labels := fs.String("labels", "", "comma-separated list of labels to apply (required)")
	comment := fs.String("comment", "", "comment body text (mutually exclusive with --comment-file)")
	commentFile := fs.String("comment-file", "", "file path or - (stdin) to read comment body from (mutually exclusive with --comment)")
	dryRun := fs.Bool("dry-run", false, "print what would be done without calling gh")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	positional := fs.Args()
	if len(positional) < 2 {
		fmt.Fprintln(stderr, "usage: gh-triage-apply [--dry-run] --labels <label,...> (--comment <text> | --comment-file <path|->) <owner/repo> <issue-number> [<issue-number>...]")
		return 2
	}

	ownerRepo := positional[0]

	// Parse and validate all issue numbers up front so we fail fast on bad input.
	issueNums := make([]int, 0, len(positional)-1)
	for _, raw := range positional[1:] {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			fmt.Fprintf(stderr, "gh-triage-apply: invalid issue number %q\n", raw)
			return 2
		}
		issueNums = append(issueNums, n)
	}

	if strings.TrimSpace(*labels) == "" {
		fmt.Fprintln(stderr, "gh-triage-apply: --labels is required and must not be empty")
		return 2
	}

	if *comment == "" && *commentFile == "" {
		fmt.Fprintln(stderr, "gh-triage-apply: one of --comment or --comment-file is required")
		return 2
	}
	if *comment != "" && *commentFile != "" {
		fmt.Fprintln(stderr, "gh-triage-apply: --comment and --comment-file are mutually exclusive")
		return 2
	}

	// Read comment body once — reused for all issues.
	body, err := readCommentBody(*comment, *commentFile, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "gh-triage-apply: %v\n", err)
		return 1
	}
	if strings.TrimSpace(body) == "" {
		fmt.Fprintln(stderr, "gh-triage-apply: comment body must not be empty")
		return 1
	}

	if *dryRun {
		for _, n := range issueNums {
			fmt.Fprintf(stdout, "[dry-run] gh issue edit %d --repo %s --add-label %q\n", n, ownerRepo, *labels)
			fmt.Fprintf(stdout, "[dry-run] gh issue comment %d --repo %s --body %q\n", n, ownerRepo, body)
		}
		return 0
	}

	// Resolve gh binary once.
	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		fmt.Fprintln(stderr, "gh-triage-apply: gh CLI not found in PATH")
		return 1
	}

	// Process all issues sequentially; collect errors (fail-aggregate).
	var errs []string
	for _, n := range issueNums {
		numStr := strconv.Itoa(n)
		if err := ghExec(ghBin, stdout, stderr, "issue", "edit",
			numStr, "--repo", ownerRepo, "--add-label", *labels,
		); err != nil {
			errs = append(errs, fmt.Sprintf("issue %d: label step failed: %v", n, err))
			continue
		}
		if err := ghExec(ghBin, stdout, stderr, "issue", "comment",
			numStr, "--repo", ownerRepo, "--body", body,
		); err != nil {
			errs = append(errs, fmt.Sprintf("issue %d: comment step failed: %v", n, err))
		}
	}

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(stderr, "gh-triage-apply: %s\n", e)
		}
		return 1
	}
	return 0
}

// readCommentBody returns the comment body from inline text, a file path, or stdin ("-").
func readCommentBody(inline, filePath string, stdin io.Reader) (string, error) {
	if inline != "" {
		return inline, nil
	}
	if filePath == "-" {
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("reading comment body from stdin: %w", err)
		}
		return string(b), nil
	}
	b, err := os.ReadFile(filePath) // #nosec G304 — filePath is user-supplied, intentional
	if err != nil {
		return "", fmt.Errorf("reading comment body from %q: %w", filePath, err)
	}
	return string(b), nil
}

// ghExec runs a gh subcommand routing output to the provided writers.
func ghExec(ghBin string, stdout, stderr io.Writer, args ...string) error {
	cmd := exec.Command(ghBin, args...) // #nosec G204 — ghBin is a resolved binary path
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
