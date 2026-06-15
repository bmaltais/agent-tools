// patch-verify applies a literal string replacement to a file and prints a
// unified diff. Exits non-zero if the pattern is not found (unlike sed).
//
// Usage: patch-verify [--all] [--dry-run] <file|- > <old-string> <new-string>
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("patch-verify", flag.ContinueOnError)
	fs.SetOutput(stderr)
	replaceAll := fs.Bool("all", false, "replace all occurrences instead of requiring exactly one")
	dryRun := fs.Bool("dry-run", false, "print diff without modifying the target file")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "patch-verify: %v\n", err)
		return 2
	}

	parts := fs.Args()
	if len(parts) != 3 {
		fmt.Fprintln(stderr, "usage: patch-verify [--all] [--dry-run] <file|- > <old-string> <new-string>")
		return 2
	}

	target := parts[0]
	oldStr := parts[1]
	newStr := parts[2]

	if oldStr == "" {
		fmt.Fprintln(stderr, "patch-verify: old-string must not be empty")
		return 1
	}

	before, err := readInput(target, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "patch-verify: %v\n", err)
		return 1
	}

	occur := strings.Count(before, oldStr)
	if occur == 0 {
		fmt.Fprintf(stderr, "patch-verify: pattern not found in %s\n", printableTarget(target))
		return 1
	}
	if occur > 1 && !*replaceAll {
		fmt.Fprintf(stderr, "patch-verify: pattern matched %d times in %s; use --all to replace all occurrences\n", occur, printableTarget(target))
		return 1
	}

	after := before
	if *replaceAll {
		after = strings.ReplaceAll(before, oldStr, newStr)
	} else {
		after = strings.Replace(before, oldStr, newStr, 1)
	}

	diff, err := buildUnifiedDiff(before, after, printableTarget(target))
	if err != nil {
		fmt.Fprintf(stderr, "patch-verify: failed to build diff: %v\n", err)
		return 1
	}

	if isTerminal(stdout) {
		diff = colorizeDiff(diff)
	}
	if _, err := io.WriteString(stdout, diff); err != nil {
		fmt.Fprintf(stderr, "patch-verify: write failed: %v\n", err)
		return 1
	}

	if target == "-" {
		if _, err := io.WriteString(stdout, "\n"); err != nil {
			fmt.Fprintf(stderr, "patch-verify: write failed: %v\n", err)
			return 1
		}
		if _, err := io.WriteString(stdout, after); err != nil {
			fmt.Fprintf(stderr, "patch-verify: write failed: %v\n", err)
			return 1
		}
		return 0
	}

	if *dryRun {
		return 0
	}

	fi, err := os.Stat(target)
	if err != nil {
		fmt.Fprintf(stderr, "patch-verify: failed to stat %s: %v\n", target, err)
		return 1
	}

	if err := os.WriteFile(target, []byte(after), fi.Mode()); err != nil {
		fmt.Fprintf(stderr, "patch-verify: failed to write %s: %v\n", target, err)
		return 1
	}

	return 0
}

func readInput(target string, stdin io.Reader) (string, error) {
	if target == "-" {
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return string(b), nil
	}

	b, err := os.ReadFile(target)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", target, err)
	}
	return string(b), nil
}

func buildUnifiedDiff(before string, after string, target string) (string, error) {
	fromName := target + " (before)"
	toName := target + " (after)"

	return difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(before),
		B:        difflib.SplitLines(after),
		FromFile: fromName,
		ToFile:   toName,
		Context:  3,
	})
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func colorizeDiff(diff string) string {
	var out bytes.Buffer
	for _, line := range strings.SplitAfter(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "@@"):
			out.WriteString("\x1b[36m")
			out.WriteString(line)
			out.WriteString("\x1b[0m")
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			out.WriteString("\x1b[32m")
			out.WriteString(line)
			out.WriteString("\x1b[0m")
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			out.WriteString("\x1b[31m")
			out.WriteString(line)
			out.WriteString("\x1b[0m")
		default:
			out.WriteString(line)
		}
	}
	return out.String()
}

func printableTarget(target string) string {
	if target == "-" {
		return "stdin"
	}
	return target
}
