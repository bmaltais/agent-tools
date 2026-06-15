// gh-merge-wait squash-merges a PR and polls until merged, with automatic
// retry on 502 and REST API fallback.
//
// Usage: gh-merge-wait <owner/repo> <pr-number> [--timeout 60s]
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "gh-merge-wait: not yet implemented — see https://github.com/bmaltais/agent-tools/issues/1")
	os.Exit(1)
}
