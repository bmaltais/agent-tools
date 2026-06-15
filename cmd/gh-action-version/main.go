// gh-action-version queries GitHub Releases for the latest tag of each action
// that targets a given Node.js runtime (e.g. node24).
//
// Usage: gh-action-version <runtime> <owner/action> [<owner/action>...]
// Example: gh-action-version node24 actions/checkout actions/setup-go softprops/action-gh-release
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "gh-action-version: not yet implemented — see https://github.com/bmaltais/agent-tools/issues/2")
	os.Exit(1)
}
