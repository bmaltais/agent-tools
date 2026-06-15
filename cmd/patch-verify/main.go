// patch-verify applies a literal string replacement to a file and prints a
// unified diff. Exits non-zero if the pattern is not found (unlike sed).
//
// Usage: patch-verify <file> <old-string> <new-string>
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "patch-verify: not yet implemented — see https://github.com/bmaltais/agent-tools/issues/4")
	os.Exit(1)
}
