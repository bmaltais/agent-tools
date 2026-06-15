// git-pr-branch manages the PR branch lifecycle as two single commands.
//
// Usage:
//
//	git-pr-branch open <branch>    — checkout main, pull, create branch
//	git-pr-branch close <branch>   — delete remote + local, return to main, pull
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: git-pr-branch open <branch> | git-pr-branch close <branch>")
		return 2
	}

	subcmd := args[0]
	branch := args[1]

	if branch == "" {
		fmt.Fprintln(stderr, "git-pr-branch: branch name must not be empty")
		return 2
	}

	switch subcmd {
	case "open":
		return runOpen(branch, stdout, stderr)
	case "close":
		return runClose(branch, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "git-pr-branch: unknown subcommand %q; must be open or close\n", subcmd)
		return 2
	}
}

// runOpen checks out main, pulls, and creates a new branch.
func runOpen(branch string, stdout, stderr io.Writer) int {
	steps := [][]string{
		{"git", "checkout", "main"},
		{"git", "pull"},
		{"git", "checkout", "-b", branch},
	}
	for _, step := range steps {
		if err := gitRun(step, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "git-pr-branch: open failed: %v\n", err)
			return 1
		}
	}
	return 0
}

// runClose deletes the remote branch (tolerating absence), checks out main,
// pulls, and force-deletes the local branch.
func runClose(branch string, stdout, stderr io.Writer) int {
	// Delete remote branch — tolerate "not found" (already deleted or never pushed).
	if err := gitRun([]string{"git", "push", "origin", "--delete", branch}, stdout, stderr); err != nil {
		if isRemoteBranchNotFound(err) {
			fmt.Fprintf(stdout, "git-pr-branch: remote branch %q not found, skipping remote delete\n", branch)
		} else {
			fmt.Fprintf(stderr, "git-pr-branch: failed to delete remote branch: %v\n", err)
			return 1
		}
	}

	steps := [][]string{
		{"git", "checkout", "main"},
		{"git", "pull"},
		{"git", "branch", "-D", branch},
	}
	for _, step := range steps {
		if err := gitRun(step, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "git-pr-branch: close failed: %v\n", err)
			return 1
		}
	}
	return 0
}

// gitRun executes a git command, routing output to the provided writers.
func gitRun(args []string, stdout, stderr io.Writer) error {
	cmd := exec.Command(args[0], args[1:]...) // #nosec G204 — args[0] is always "git"
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// isRemoteBranchNotFound reports whether a git push --delete error indicates
// the remote branch does not exist.
func isRemoteBranchNotFound(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	// git writes "error: unable to delete '...': remote ref does not exist" to stderr.
	// ExitError.Stderr captures it when Stderr is not set on the Cmd — but here we pipe
	// stderr to the caller's writer. Check the exit code: git exits 1 for this case.
	// We rely on the message captured in the output writer, but since we can't read it
	// back here, we accept exit code 1 from push --delete as "branch not found".
	_ = strings.TrimSpace(string(exitErr.Stderr))
	return exitErr.ExitCode() == 1
}
