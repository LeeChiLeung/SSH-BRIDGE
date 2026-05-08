package main

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

func TestCurrentBuildInfoNormalizesVersion(t *testing.T) {
	originalVersion := buildVersion
	originalCommit := buildCommit
	originalDate := buildDate
	t.Cleanup(func() {
		buildVersion = originalVersion
		buildCommit = originalCommit
		buildDate = originalDate
	})

	buildVersion = "v0.0.2"
	buildCommit = "abcdef1234567890"
	buildDate = "2026-05-08T08:00:00Z"

	info := currentBuildInfo()
	if info.Version != "0.0.2" {
		t.Fatalf("unexpected normalized version: %q", info.Version)
	}
	if info.Commit != "abcdef1234567890" {
		t.Fatalf("unexpected commit: %q", info.Commit)
	}
	if info.Date != "2026-05-08T08:00:00Z" {
		t.Fatalf("unexpected date: %q", info.Date)
	}
}

func TestVersionStringIncludesBuildMetadata(t *testing.T) {
	originalVersion := buildVersion
	originalCommit := buildCommit
	originalDate := buildDate
	t.Cleanup(func() {
		buildVersion = originalVersion
		buildCommit = originalCommit
		buildDate = originalDate
	})

	buildVersion = "0.1.0"
	buildCommit = "abc123"
	buildDate = "2026-05-08T08:00:00Z"

	output := versionString()
	if !strings.Contains(output, "0.1.0") || !strings.Contains(output, "abc123") || !strings.Contains(output, "2026-05-08T08:00:00Z") {
		t.Fatalf("unexpected version string: %q", output)
	}
}

func TestWebTemplateShowsVersionFooter(t *testing.T) {
	tpls, err := template.New("page").Funcs(template.FuncMap{
		"formatTime":  formatTime,
		"formatBytes": formatBytes,
		"formatRate":  formatRate,
		"statusLabel": func(state TunnelState) string { return string(state) },
	}).Parse(webTemplates)
	if err != nil {
		t.Fatalf("parse web templates: %v", err)
	}

	var buf bytes.Buffer
	err = tpls.ExecuteTemplate(&buf, "page", indexPageData{
		PageTitle: "SSH Tunnel Console",
		Version:   "0.0.2",
	})
	if err != nil {
		t.Fatalf("execute page template: %v", err)
	}
	if !strings.Contains(buf.String(), "Version 0.0.2") {
		t.Fatalf("page footer missing version: %s", buf.String())
	}
}
