// gh-action-version queries GitHub Releases for the latest tag of each action
// that targets a given Node.js runtime (e.g. node24).
//
// Usage: gh-action-version [--output yaml] <runtime> <owner/action> [<owner/action>...]
// Example: gh-action-version node24 actions/checkout actions/setup-go softprops/action-gh-release
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/cli/safeexec"
)

// output format values.
const (
	outputText = "text"
	outputYAML = "yaml"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gh-action-version", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outputFmt := fs.String("output", outputText, "output format: text or yaml")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	positional := fs.Args()
	if len(positional) < 2 {
		fmt.Fprintln(stderr, "usage: gh-action-version [--output yaml] <runtime> <owner/action> [<owner/action>...]")
		return 2
	}

	if *outputFmt != outputText && *outputFmt != outputYAML {
		fmt.Fprintf(stderr, "gh-action-version: invalid --output %q; must be text or yaml\n", *outputFmt)
		return 2
	}

	runtime := positional[0]
	actions := positional[1:]

	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		fmt.Fprintln(stderr, "gh-action-version: gh CLI not found in PATH")
		return 1
	}

	results := resolveParallel(ghBin, runtime, actions)

	// Check for any errors before printing.
	hasError := false
	for _, r := range results {
		if r.err != nil {
			fmt.Fprintf(stderr, "gh-action-version: %s: %v\n", r.action, r.err)
			hasError = true
		}
	}
	if hasError {
		return 1
	}

	printResults(stdout, results, *outputFmt)
	return 0
}

// result holds the resolved tag for one action, or an error.
type result struct {
	action string
	tag    string
	err    error
}

// resolveParallel looks up the latest matching release for each action concurrently.
func resolveParallel(ghBin, runtime string, actions []string) []result {
	results := make([]result, len(actions))
	var wg sync.WaitGroup
	for i, action := range actions {
		wg.Add(1)
		go func(idx int, act string) {
			defer wg.Done()
			tag, err := resolveAction(ghBin, runtime, act)
			results[idx] = result{action: act, tag: tag, err: err}
		}(i, action)
	}
	wg.Wait()
	return results
}

// resolveAction finds the latest GitHub release for the action whose body
// contains the runtime string (case-insensitive).
func resolveAction(ghBin, runtime, action string) (string, error) {
	path := fmt.Sprintf("repos/%s/releases", action)
	data, err := ghAPI(ghBin, path)
	if err != nil {
		return "", fmt.Errorf("fetch releases: %w", err)
	}

	var releases []release
	if err := json.Unmarshal(data, &releases); err != nil {
		return "", fmt.Errorf("parse releases: %w", err)
	}

	runtimeLower := strings.ToLower(runtime)
	for _, rel := range releases {
		if strings.Contains(strings.ToLower(rel.Body), runtimeLower) {
			return rel.TagName, nil
		}
	}
	return "", fmt.Errorf("no release found targeting runtime %q", runtime)
}

// release is the subset of the GitHub releases API response we need.
type release struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
}

// printResults writes results to stdout in the requested format.
func printResults(stdout io.Writer, results []result, format string) {
	for _, r := range results {
		switch format {
		case outputYAML:
			fmt.Fprintf(stdout, "%s: %s\n", r.action, r.tag)
		default:
			fmt.Fprintf(stdout, "%s@%s\n", r.action, r.tag)
		}
	}
}

// ghAPI runs `gh api <path>` and returns the response body.
func ghAPI(ghBin, path string) ([]byte, error) {
	var out, errOut bytes.Buffer
	cmd := exec.Command(ghBin, "api", path) // #nosec G204
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(errOut.String()))
	}
	return out.Bytes(), nil
}
