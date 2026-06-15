// gh-pr-comments fetches both inline review comments and conversation-level
// comments for a GitHub pull request and returns a unified JSON array.
//
// Usage: gh-pr-comments <owner/repo> <pr-number>
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cli/safeexec"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: gh-pr-comments <owner/repo> <pr-number>")
		return 2
	}

	ownerRepo := args[0]
	prNum, err := strconv.Atoi(args[1])
	if err != nil || prNum < 1 {
		fmt.Fprintf(stderr, "gh-pr-comments: invalid PR number %q\n", args[1])
		return 2
	}

	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		fmt.Fprintln(stderr, "gh-pr-comments: gh CLI not found in PATH")
		return 1
	}

	comments, err := fetchAllComments(ghBin, ownerRepo, prNum)
	if err != nil {
		fmt.Fprintf(stderr, "gh-pr-comments: %v\n", err)
		return 1
	}

	out, err := json.Marshal(comments)
	if err != nil {
		fmt.Fprintf(stderr, "gh-pr-comments: failed to encode output: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, string(out))
	return 0
}

// comment type discriminators.
const (
	typeInline       = "inline"
	typeConversation = "conversation"
)

// comment is the unified output element.
type comment struct {
	Type   string `json:"type"`
	ID     int64  `json:"id"`
	Path   string `json:"path,omitempty"` // inline only
	Line   int    `json:"line,omitempty"` // inline only
	Body   string `json:"body"`
	Author string `json:"author"`
}

// inlineAPIComment represents the GitHub API response for a pull request review comment.
type inlineAPIComment struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
	Line int    `json:"line"`
	Body string `json:"body"`
	User struct {
		Login string `json:"login"`
	} `json:"user"`
}

// conversationAPIComment represents the GitHub API response for an issue comment.
type conversationAPIComment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
	User struct {
		Login string `json:"login"`
	} `json:"user"`
}

// fetchAllComments retrieves both inline and conversation comments and merges them.
func fetchAllComments(ghBin, ownerRepo string, prNum int) ([]comment, error) {
	inline, err := fetchInlineComments(ghBin, ownerRepo, prNum)
	if err != nil {
		return nil, fmt.Errorf("inline comments: %w", err)
	}

	conversation, err := fetchConversationComments(ghBin, ownerRepo, prNum)
	if err != nil {
		return nil, fmt.Errorf("conversation comments: %w", err)
	}

	result := make([]comment, 0, len(inline)+len(conversation))
	result = append(result, inline...)
	result = append(result, conversation...)
	return result, nil
}

// fetchInlineComments calls GET /repos/{owner}/{repo}/pulls/{N}/comments.
func fetchInlineComments(ghBin, ownerRepo string, prNum int) ([]comment, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d/comments", ownerRepo, prNum)
	data, err := ghAPI(ghBin, path)
	if err != nil {
		return nil, err
	}

	var raw []inlineAPIComment
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	out := make([]comment, len(raw))
	for i, c := range raw {
		out[i] = comment{
			Type:   typeInline,
			ID:     c.ID,
			Path:   c.Path,
			Line:   c.Line,
			Body:   c.Body,
			Author: c.User.Login,
		}
	}
	return out, nil
}

// fetchConversationComments calls GET /repos/{owner}/{repo}/issues/{N}/comments.
func fetchConversationComments(ghBin, ownerRepo string, prNum int) ([]comment, error) {
	path := fmt.Sprintf("repos/%s/issues/%d/comments", ownerRepo, prNum)
	data, err := ghAPI(ghBin, path)
	if err != nil {
		return nil, err
	}

	var raw []conversationAPIComment
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	out := make([]comment, len(raw))
	for i, c := range raw {
		out[i] = comment{
			Type:   typeConversation,
			ID:     c.ID,
			Body:   c.Body,
			Author: c.User.Login,
		}
	}
	return out, nil
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
