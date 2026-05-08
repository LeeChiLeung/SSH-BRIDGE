package main

import (
	"fmt"
	"runtime/debug"
	"strings"
)

var (
	buildVersion = "0.0.0-dev"
	buildCommit  = "unknown"
	buildDate    = "unknown"
)

type buildInfo struct {
	Version string
	Commit  string
	Date    string
}

func currentBuildInfo() buildInfo {
	info := buildInfo{
		Version: normalizeVersion(buildVersion),
		Commit:  strings.TrimSpace(buildCommit),
		Date:    strings.TrimSpace(buildDate),
	}

	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range bi.Settings {
			switch setting.Key {
			case "vcs.revision":
				if info.Commit == "" || info.Commit == "unknown" {
					info.Commit = shortCommit(setting.Value)
				}
			case "vcs.time":
				if info.Date == "" || info.Date == "unknown" {
					info.Date = setting.Value
				}
			}
		}
	}

	if info.Commit == "" {
		info.Commit = "unknown"
	}
	if info.Date == "" {
		info.Date = "unknown"
	}
	return info
}

func normalizeVersion(raw string) string {
	version := strings.TrimSpace(raw)
	if version == "" {
		return "0.0.0-dev"
	}
	return strings.TrimPrefix(version, "v")
}

func shortCommit(raw string) string {
	commit := strings.TrimSpace(raw)
	if len(commit) > 12 {
		return commit[:12]
	}
	if commit == "" {
		return "unknown"
	}
	return commit
}

func versionString() string {
	info := currentBuildInfo()
	if info.Commit == "unknown" && info.Date == "unknown" {
		return info.Version
	}
	return fmt.Sprintf("%s (commit %s, built %s)", info.Version, info.Commit, info.Date)
}
