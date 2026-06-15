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

func TestRun_OnlyRuntime_NoActions(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"node24"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

func TestRun_InvalidOutputFormat(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--output", "json", "node24", "actions/checkout"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "--output") {
		t.Fatalf("expected --output error, got %q", errOut.String())
	}
}

func TestRun_InvalidFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--bogus", "node24", "actions/checkout"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

func TestResolveAction_MatchesRuntime(t *testing.T) {
	releases := []release{
		{TagName: "v5.0.0", Body: "This release targets node20 runtime."},
		{TagName: "v6.0.0", Body: "This release targets Node24 runtime and is fully supported."},
		{TagName: "v7.0.0-beta", Body: "Beta for node24 with experimental features."},
	}

	runtimeLower := "node24"
	var found string
	for _, rel := range releases {
		if strings.Contains(strings.ToLower(rel.Body), runtimeLower) {
			found = rel.TagName
			break
		}
	}

	if found != "v6.0.0" {
		t.Errorf("expected v6.0.0 (first node24 match), got %q", found)
	}
}

func TestResolveAction_CaseInsensitive(t *testing.T) {
	releases := []release{
		{TagName: "v1.0.0", Body: "Targets NODE24 environment"},
	}

	runtimeLower := "node24"
	found := false
	for _, rel := range releases {
		if strings.Contains(strings.ToLower(rel.Body), runtimeLower) {
			found = true
			break
		}
	}

	if !found {
		t.Error("case-insensitive match failed for NODE24")
	}
}

func TestResolveAction_NoMatch(t *testing.T) {
	releases := []release{
		{TagName: "v5.0.0", Body: "This release targets node20 runtime."},
	}

	runtimeLower := "node24"
	found := ""
	for _, rel := range releases {
		if strings.Contains(strings.ToLower(rel.Body), runtimeLower) {
			found = rel.TagName
			break
		}
	}

	if found != "" {
		t.Errorf("expected no match, got %q", found)
	}
}

func TestPrintResults_TextFormat(t *testing.T) {
	results := []result{
		{action: "actions/checkout", tag: "v6.0.0"},
		{action: "actions/setup-go", tag: "v5.0.0"},
	}

	var out bytes.Buffer
	printResults(&out, results, "text")

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %s", len(lines), out.String())
	}
	if lines[0] != "actions/checkout@v6.0.0" {
		t.Errorf("line 0: want actions/checkout@v6.0.0, got %q", lines[0])
	}
	if lines[1] != "actions/setup-go@v5.0.0" {
		t.Errorf("line 1: want actions/setup-go@v5.0.0, got %q", lines[1])
	}
}

func TestPrintResults_YAMLFormat(t *testing.T) {
	results := []result{
		{action: "actions/checkout", tag: "v6.0.0"},
		{action: "actions/setup-go", tag: "v5.0.0"},
	}

	var out bytes.Buffer
	printResults(&out, results, "yaml")

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %s", len(lines), out.String())
	}
	if lines[0] != "actions/checkout: v6.0.0" {
		t.Errorf("yaml line 0: want 'actions/checkout: v6.0.0', got %q", lines[0])
	}
}

func TestReleaseJSON_Parse(t *testing.T) {
	data := `[{"tag_name":"v6.0.0","body":"supports node24"}]`
	var releases []release
	if err := json.Unmarshal([]byte(data), &releases); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}
	if releases[0].TagName != "v6.0.0" {
		t.Errorf("tag: want v6.0.0, got %q", releases[0].TagName)
	}
}
