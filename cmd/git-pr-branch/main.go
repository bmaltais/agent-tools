// git-pr-branch manages the branch lifecycle for a PR.
//
// Usage:
//   git-pr-branch open <branch>    — checkout main, pull, create branch
//   git-pr-branch close <branch>   — delete remote + local branch, return to main, pull
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "git-pr-branch: not yet implemented — see https://github.com/bmaltais/agent-tools/issues/3")
	os.Exit(1)
}
