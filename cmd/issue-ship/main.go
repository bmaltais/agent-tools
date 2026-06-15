// issue-ship drives the full triage → branch → PR → merge → cleanup pipeline
// for a GitHub issue, resuming from wherever it last stopped.
//
// Usage: issue-ship <owner/repo> <issue-number> [--method squash]
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "issue-ship: not yet implemented — see https://github.com/bmaltais/agent-tools/issues/5")
	os.Exit(1)
}
