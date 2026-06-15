// issue-ship drives the full agent issue pipeline — branch → PR → merge → cleanup
// as a single idempotent command. It detects the current stage and resumes from there.
//
// Usage: issue-ship [--dry-run] [--method squash|rebase|merge] [--from-stage branch|pr|merge|cleanup] <owner/repo> <issue-number>
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cli/safeexec"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// stages defines the ordered pipeline stages.
var stages = []string{"branch", "pr", "merge", "cleanup"}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("issue-ship", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dryRun := fs.Bool("dry-run", false, "print each stage action without executing")
	method := fs.String("method", "squash", "merge method: squash, rebase, or merge")
	fromStage := fs.String("from-stage", "", "force restart from a specific stage (branch|pr|merge|cleanup)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	positional := fs.Args()
	if len(positional) != 2 {
		fmt.Fprintln(stderr, "usage: issue-ship [--dry-run] [--method squash|rebase|merge] [--from-stage <stage>] <owner/repo> <issue-number>")
		return 2
	}

	ownerRepo := positional[0]
	issueNum, err := strconv.Atoi(positional[1])
	if err != nil || issueNum < 1 {
		fmt.Fprintf(stderr, "issue-ship: invalid issue number %q\n", positional[1])
		return 2
	}

	validMethods := map[string]bool{"squash": true, "rebase": true, "merge": true}
	if !validMethods[*method] {
		fmt.Fprintf(stderr, "issue-ship: invalid --method %q; must be squash, rebase, or merge\n", *method)
		return 2
	}

	if *fromStage != "" && !validStage(*fromStage) {
		fmt.Fprintf(stderr, "issue-ship: invalid --from-stage %q; must be branch, pr, merge, or cleanup\n", *fromStage)
		return 2
	}

	// Resolve gh binary once.
	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		fmt.Fprintln(stderr, "issue-ship: gh CLI not found in PATH")
		return 1
	}

	// Fetch issue title to construct branch name.
	title, err := issueTitle(ghBin, ownerRepo, issueNum)
	if err != nil {
		fmt.Fprintf(stderr, "issue-ship: failed to fetch issue title: %v\n", err)
		return 1
	}

	branch := branchName(issueNum, title)

	cfg := &config{
		ghBin:     ghBin,
		ownerRepo: ownerRepo,
		issueNum:  issueNum,
		branch:    branch,
		method:    *method,
		dryRun:    *dryRun,
		stdout:    stdout,
		stderr:    stderr,
	}

	startStage := *fromStage
	if startStage == "" {
		startStage, err = detectStage(cfg)
		if err != nil {
			fmt.Fprintf(stderr, "issue-ship: stage detection failed: %v\n", err)
			return 1
		}
	}

	if startStage == "" {
		fmt.Fprintf(stdout, "issue-ship: pipeline already complete for issue #%d\n", issueNum)
		return 0
	}

	return runFrom(cfg, startStage)
}

type config struct {
	ghBin     string
	ownerRepo string
	issueNum  int
	branch    string
	method    string
	dryRun    bool
	stdout    io.Writer
	stderr    io.Writer
}

// detectStage inspects the current state and returns the stage to start from,
// or "" if the pipeline is already complete.
func detectStage(cfg *config) (string, error) {
	// Check if remote branch exists.
	branchExists, err := remoteBranchExists(cfg)
	if err != nil {
		return "", fmt.Errorf("checking branch: %w", err)
	}
	if !branchExists {
		return "branch", nil
	}

	// Branch exists — check for open PR.
	prNum, err := openPRForBranch(cfg)
	if err != nil {
		return "", fmt.Errorf("checking open PR: %w", err)
	}
	if prNum > 0 {
		return "merge", nil
	}

	// No open PR — check for merged PR.
	merged, err := mergedPRForBranch(cfg)
	if err != nil {
		return "", fmt.Errorf("checking merged PR: %w", err)
	}
	if merged {
		// Branch still exists after merge → cleanup.
		return "cleanup", nil
	}

	// Branch exists but no PR at all → PR was not yet opened.
	return "pr", nil
}

// runFrom executes the pipeline starting at the named stage.
func runFrom(cfg *config, startStage string) int {
	idx := stageIndex(startStage)
	for _, stage := range stages[idx:] {
		var err error
		switch stage {
		case "branch":
			err = runBranch(cfg)
		case "pr":
			err = runPR(cfg)
		case "merge":
			err = runMerge(cfg)
		case "cleanup":
			err = runCleanup(cfg)
		}
		if err != nil {
			fmt.Fprintf(cfg.stderr, "issue-ship: [%s] failed: %v\n", stage, err)
			return 1
		}
	}
	return 0
}

func runBranch(cfg *config) error {
	cfg.logStage("branch", "create "+cfg.branch+" from main")
	if cfg.dryRun {
		return nil
	}
	// Ensure we are on main and up to date.
	if err := ghRun(cfg, "git", "checkout", "main"); err != nil {
		return err
	}
	if err := ghRun(cfg, "git", "pull"); err != nil {
		return err
	}
	return ghRun(cfg, "git", "checkout", "-b", cfg.branch)
}

func runPR(cfg *config) error {
	body := fmt.Sprintf("Closes #%d", cfg.issueNum)
	cfg.logStage("pr", fmt.Sprintf("open PR for %s", cfg.branch))
	if cfg.dryRun {
		return nil
	}
	return ghRun(cfg, cfg.ghBin, "pr", "create",
		"--repo", cfg.ownerRepo,
		"--head", cfg.branch,
		"--base", "main",
		"--title", fmt.Sprintf("feat: implement issue #%d", cfg.issueNum),
		"--body", body,
	)
}

func runMerge(cfg *config) error {
	cfg.logStage("merge", fmt.Sprintf("%s-merge PR for %s", cfg.method, cfg.branch))
	if cfg.dryRun {
		return nil
	}
	return ghRun(cfg, cfg.ghBin, "pr", "merge",
		"--repo", cfg.ownerRepo,
		"--head", cfg.branch,
		"--"+cfg.method,
		"--auto",
	)
}

func runCleanup(cfg *config) error {
	cfg.logStage("cleanup", "delete remote branch and return to main")
	if cfg.dryRun {
		return nil
	}
	if err := ghRun(cfg, cfg.ghBin, "api",
		"--method", "DELETE",
		fmt.Sprintf("repos/%s/git/refs/heads/%s", cfg.ownerRepo, cfg.branch),
	); err != nil {
		// Non-fatal if branch already gone.
		fmt.Fprintf(cfg.stderr, "issue-ship: [cleanup] warning: could not delete remote branch: %v\n", err)
	}
	if err := ghRun(cfg, "git", "checkout", "main"); err != nil {
		return err
	}
	return ghRun(cfg, "git", "pull")
}

func (cfg *config) logStage(stage, action string) {
	prefix := ""
	if cfg.dryRun {
		prefix = "[dry-run] "
	}
	fmt.Fprintf(cfg.stdout, "%s[%s] %s\n", prefix, stage, action)
}

// ghRun executes a command, routing stdout/stderr to the config writers.
func ghRun(cfg *config, name string, args ...string) error {
	cmd := exec.Command(name, args...) // #nosec G204 — name is always a resolved binary path or "git"
	cmd.Stdout = cfg.stdout
	cmd.Stderr = cfg.stderr
	return cmd.Run()
}

// remoteBranchExists returns true if the remote branch already exists via gh api.
func remoteBranchExists(cfg *config) (bool, error) {
	ghBin := cfg.ghBin
	path := fmt.Sprintf("repos/%s/git/refs/heads/%s", cfg.ownerRepo, cfg.branch)
	cmd := exec.Command(ghBin, "api", path) // #nosec G204
	cmd.Stderr = io.Discard
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		// gh api exits 1 on 404.
		return false, nil
	}
	return false, err
}

// openPRForBranch returns the PR number if an open PR exists for the branch, else 0.
func openPRForBranch(cfg *config) (int, error) {
	parts := strings.SplitN(cfg.ownerRepo, "/", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid owner/repo: %s", cfg.ownerRepo)
	}
	path := fmt.Sprintf("repos/%s/pulls?head=%s:%s&state=open&per_page=1",
		cfg.ownerRepo, parts[0], cfg.branch)
	cmd := exec.Command(cfg.ghBin, "api", path) // #nosec G204
	cmd.Stderr = cfg.stderr
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var prs []struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(out, &prs); err != nil {
		return 0, err
	}
	if len(prs) > 0 {
		return prs[0].Number, nil
	}
	return 0, nil
}

// mergedPRForBranch returns true if a closed+merged PR exists for the branch.
func mergedPRForBranch(cfg *config) (bool, error) {
	parts := strings.SplitN(cfg.ownerRepo, "/", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid owner/repo: %s", cfg.ownerRepo)
	}
	path := fmt.Sprintf("repos/%s/pulls?head=%s:%s&state=closed&per_page=10",
		cfg.ownerRepo, parts[0], cfg.branch)
	cmd := exec.Command(cfg.ghBin, "api", path) // #nosec G204
	cmd.Stderr = cfg.stderr
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	var prs []struct {
		MergedAt *string `json:"merged_at"`
	}
	if err := json.Unmarshal(out, &prs); err != nil {
		return false, err
	}
	for _, pr := range prs {
		if pr.MergedAt != nil {
			return true, nil
		}
	}
	return false, nil
}

// issueTitle fetches the title of an issue.
func issueTitle(ghBin, ownerRepo string, issueNum int) (string, error) {
	path := fmt.Sprintf("repos/%s/issues/%d", ownerRepo, issueNum)
	cmd := exec.Command(ghBin, "api", path) // #nosec G204
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	var issue struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(out, &issue); err != nil {
		return "", err
	}
	return issue.Title, nil
}

// slugify returns a kebab-case slug from the first 5 words of title.
var nonAlphanumSpace = regexp.MustCompile(`[^a-z0-9 ]+`)

func slugify(title string) string {
	lower := strings.ToLower(title)
	cleaned := nonAlphanumSpace.ReplaceAllString(lower, "")
	words := strings.Fields(cleaned)
	if len(words) > 5 {
		words = words[:5]
	}
	return strings.Join(words, "-")
}

// branchName returns the canonical branch name for an issue.
func branchName(issueNum int, title string) string {
	return fmt.Sprintf("feat/issue-%d-%s", issueNum, slugify(title))
}

func validStage(s string) bool {
	for _, stage := range stages {
		if stage == s {
			return true
		}
	}
	return false
}

func stageIndex(s string) int {
	for i, stage := range stages {
		if stage == s {
			return i
		}
	}
	return 0
}


