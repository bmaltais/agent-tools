package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRun_MissingArgs(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "usage:") {
		t.Fatalf("expected usage message, got %q", errOut.String())
	}
}

func TestRun_TooManyArgs(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"owner/repo", "1", "extra"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

func TestRun_InvalidPRNumber(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"owner/repo", "abc"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "invalid PR number") {
		t.Fatalf("expected invalid PR number message, got %q", errOut.String())
	}
}

func TestRun_ZeroPRNumber(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"owner/repo", "0"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

func TestParseInlineComments_Shape(t *testing.T) {
	raw := []inlineAPIComment{
		{ID: 1, Path: "main.go", Line: 42, Body: "fix this", User: struct{ Login string `json:"login"` }{Login: "alice"}},
	}

	b, _ := json.Marshal(raw)
	var parsed []inlineAPIComment
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	c := comment{
		Type:   "inline",
		ID:     parsed[0].ID,
		Path:   parsed[0].Path,
		Line:   parsed[0].Line,
		Body:   parsed[0].Body,
		Author: parsed[0].User.Login,
	}

	if c.Type != "inline" {
		t.Errorf("type: want inline, got %q", c.Type)
	}
	if c.Path != "main.go" {
		t.Errorf("path: want main.go, got %q", c.Path)
	}
	if c.Line != 42 {
		t.Errorf("line: want 42, got %d", c.Line)
	}
	if c.Author != "alice" {
		t.Errorf("author: want alice, got %q", c.Author)
	}
}

func TestParseConversationComments_Shape(t *testing.T) {
	raw := []conversationAPIComment{
		{ID: 99, Body: "lgtm", User: struct{ Login string `json:"login"` }{Login: "bob"}},
	}

	b, _ := json.Marshal(raw)
	var parsed []conversationAPIComment
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	c := comment{
		Type:   "conversation",
		ID:     parsed[0].ID,
		Body:   parsed[0].Body,
		Author: parsed[0].User.Login,
	}

	if c.Type != "conversation" {
		t.Errorf("type: want conversation, got %q", c.Type)
	}
	if c.Path != "" {
		t.Errorf("path should be empty for conversation, got %q", c.Path)
	}
	if c.Author != "bob" {
		t.Errorf("author: want bob, got %q", c.Author)
	}
}

func TestMarshalEmptyResult(t *testing.T) {
	comments := []comment{}
	out, err := json.Marshal(comments)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != "[]" {
		t.Errorf("expected [], got %s", out)
	}
}

func TestCommentJSONOmitsEmptyPathAndLine(t *testing.T) {
	// Conversation comments should omit path and line in JSON output.
	c := comment{Type: "conversation", ID: 1, Body: "ok", Author: "alice"}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	if strings.Contains(s, `"path"`) {
		t.Errorf("conversation comment should omit path, got: %s", s)
	}
	if strings.Contains(s, `"line"`) {
		t.Errorf("conversation comment should omit line, got: %s", s)
	}
}

func TestCommentJSONIncludesPathAndLine_Inline(t *testing.T) {
	c := comment{Type: "inline", ID: 1, Path: "foo.go", Line: 5, Body: "fix", Author: "alice"}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, `"path"`) {
		t.Errorf("inline comment should include path, got: %s", s)
	}
	if !strings.Contains(s, `"line"`) {
		t.Errorf("inline comment should include line, got: %s", s)
	}
}
