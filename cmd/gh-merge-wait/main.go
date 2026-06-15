// gh-merge-wait squash-merges a PR and polls until merged, with automatic
// retry on 502 and REST API fallback.
//
// Usage: gh-merge-wait [--timeout 60s] [--method squash|merge|rebase] <owner/repo> <pr-number>
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cli/safeexec"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gh-merge-wait", flag.ContinueOnError)
	fs.SetOutput(stderr)
	timeout := fs.Duration("timeout", 60*time.Second, "maximum time to wait for merge confirmation")
	method := fs.String("method", "squash", "merge method: squash, merge, or rebase")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	positional := fs.Args()
	if len(positional) != 2 {
		fmt.Fprintln(stderr, "usage: gh-merge-wait [--timeout 60s] [--method squash|merge|rebase] <owner/repo> <pr-number>")
		return 2
	}

	ownerRepo := positional[0]
	prNum, err := strconv.Atoi(positional[1])
	if err != nil || prNum < 1 {
		fmt.Fprintf(stderr, "gh-merge-wait: invalid PR number %q\n", positional[1])
		return 2
	}

	validMethods := map[string]bool{"squash": true, "merge": true, "rebase": true}
	if !validMethods[*method] {
		fmt.Fprintf(stderr, "gh-merge-wait: invalid --method %q; must be squash, merge, or rebase\n", *method)
		return 2
	}

	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		fmt.Fprintln(stderr, "gh-merge-wait: gh CLI not found in PATH")
		return 1
	}

	deadline := time.Now().Add(*timeout)

	if err := triggerMerge(ghBin, ownerRepo, prNum, *method, stderr); err != nil {
		fmt.Fprintf(stderr, "gh-merge-wait: failed to trigger merge: %v\n", err)
		return 1
	}

	sha, err := pollUntilMerged(ghBin, ownerRepo, prNum, deadline)
	if err != nil {
		fmt.Fprintf(stderr, "gh-merge-wait: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, sha)
	return 0
}

// triggerMerge attempts to merge via CLI, falling back to REST API on 502.
func triggerMerge(ghBin, ownerRepo string, prNum int, method string, stderr io.Writer) error {
	var cliStderr bytes.Buffer
	cmd := exec.Command(ghBin, "pr", "merge", // #nosec G204
		"--repo", ownerRepo,
		"--"+method,
		strconv.Itoa(prNum),
	)
	cmd.Stderr = &cliStderr
	if err := cmd.Run(); err != nil {
		if is502Output(cliStderr.String()) {
			fmt.Fprintln(stderr, "gh-merge-wait: CLI returned 502, retrying via REST API")
			return mergeViaREST(ghBin, ownerRepo, prNum, method)
		}
		return fmt.Errorf("CLI: %w: %s", err, strings.TrimSpace(cliStderr.String()))
	}
	return nil
}

// mergeViaREST calls PUT /repos/{owner}/{repo}/pulls/{N}/merge via gh api.
func mergeViaREST(ghBin, ownerRepo string, prNum int, method string) error {
	var errOut bytes.Buffer
	cmd := exec.Command(ghBin, "api", // #nosec G204
		"-X", "PUT",
		fmt.Sprintf("repos/%s/pulls/%d/merge", ownerRepo, prNum),
		"-f", "merge_method="+method,
	)
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("REST: %w: %s", err, strings.TrimSpace(errOut.String()))
	}
	return nil
}

// pollInterval is the delay between successive merge-state polls.
const pollInterval = 3 * time.Second

// pullResponse holds the fields we care about from GET /repos/{owner}/{repo}/pulls/{N}.
type pullResponse struct {
	Merged         bool   `json:"merged"`
	MergeCommitSHA string `json:"merge_commit_sha"`
}

// pollUntilMerged polls GET /pulls/{N} until merged==true or deadline passes.
// Returns the merge commit SHA on success.
func pollUntilMerged(ghBin, ownerRepo string, prNum int, deadline time.Time) (string, error) {
	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timeout waiting for PR #%d to be merged", prNum)
		}

		pr, err := fetchPR(ghBin, ownerRepo, prNum)
		if err != nil {
			return "", fmt.Errorf("polling: %w", err)
		}

		if pr.Merged && pr.MergeCommitSHA != "" {
			return pr.MergeCommitSHA, nil
		}

			time.Sleep(pollInterval)
	}
}

// fetchPR retrieves the pull request state from the GitHub API.
func fetchPR(ghBin, ownerRepo string, prNum int) (*pullResponse, error) {
	var out, errOut bytes.Buffer
	cmd := exec.Command(ghBin, "api", // #nosec G204
		fmt.Sprintf("repos/%s/pulls/%d", ownerRepo, prNum),
	)
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(errOut.String()))
	}

	var pr pullResponse
	if err := json.Unmarshal(out.Bytes(), &pr); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &pr, nil
}

// is502Output reports whether the output string indicates an HTTP 502 response.
func is502Output(s string) bool {
	return strings.Contains(s, "502")
}
