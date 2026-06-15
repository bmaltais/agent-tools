// gh-triage-apply applies a label set and posts a comment to a GitHub issue in a
// single command, replacing the two-call sequence of gh issue edit + gh issue comment.
//
// Usage: gh-triage-apply <owner/repo> <issue-number> --labels <label,...> (--comment <text> | --comment-file <path|->)
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
	if len(positional) != 2 {
		fmt.Fprintln(stderr, "usage: gh-triage-apply <owner/repo> <issue-number> --labels <label,...> (--comment <text> | --comment-file <path|->)")
		return 2
	}

	ownerRepo := positional[0]
	issueNum, err := strconv.Atoi(positional[1])
	if err != nil || issueNum < 1 {
		fmt.Fprintf(stderr, "gh-triage-apply: invalid issue number %q\n", positional[1])
		return 2
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

	// Read comment body.
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
		fmt.Fprintf(stdout, "[dry-run] gh issue edit %d --repo %s --add-label %q\n", issueNum, ownerRepo, *labels)
		fmt.Fprintf(stdout, "[dry-run] gh issue comment %d --repo %s --body %q\n", issueNum, ownerRepo, body)
		return 0
	}

	// Resolve gh binary.
	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		fmt.Fprintln(stderr, "gh-triage-apply: gh CLI not found in PATH")
		return 1
	}

	// Step 1: apply labels.
	if err := ghExec(ghBin, stdout, stderr, "issue", "edit",
		strconv.Itoa(issueNum), "--repo", ownerRepo, "--add-label", *labels,
	); err != nil {
		fmt.Fprintf(stderr, "gh-triage-apply: label step failed: %v\n", err)
		return 1
	}

	// Step 2: post comment.
	if err := ghExec(ghBin, stdout, stderr, "issue", "comment",
		strconv.Itoa(issueNum), "--repo", ownerRepo, "--body", body,
	); err != nil {
		fmt.Fprintf(stderr, "gh-triage-apply: comment step failed: %v\n", err)
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
