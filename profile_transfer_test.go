package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func TestProfileStoreExportTransferJSON(t *testing.T) {
	store, err := NewProfileStore(filepath.Join(t.TempDir(), "profiles.json"))
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}
	if _, err := store.Save(Profile{
		Title:             "数据库隧道",
		SSHHost:           "ssh.example.com",
		SSHPort:           22,
		SSHUser:           "root",
		SSHPassword:       "secret",
		RemoteHost:        "127.0.0.1",
		RemotePort:        3306,
		LocalAddr:         "127.0.0.1",
		LocalPort:         13306,
		ConnectTimeoutSec: 10,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	now := time.Date(2026, 5, 9, 2, 0, 0, 0, time.UTC)
	data, err := store.ExportTransferJSON(now)
	if err != nil {
		t.Fatalf("ExportTransferJSON: %v", err)
	}
	var payload ProfileTransferFile
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if payload.SchemaVersion != profileTransferSchemaVersion {
		t.Fatalf("SchemaVersion = %q", payload.SchemaVersion)
	}
	if payload.ExportedAt != now.Format(time.RFC3339) {
		t.Fatalf("ExportedAt = %q", payload.ExportedAt)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Title != "数据库隧道" {
		t.Fatalf("unexpected profiles payload: %+v", payload.Profiles)
	}
}

func TestProfileStorePreviewImportDetectsConflictsAndInvalid(t *testing.T) {
	store, err := NewProfileStore(filepath.Join(t.TempDir(), "profiles.json"))
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}
	if _, err := store.Save(Profile{
		Title:             "数据库隧道",
		SSHHost:           "ssh.example.com",
		SSHPort:           22,
		SSHUser:           "root",
		SSHKeyPath:        "~/.ssh/id_rsa",
		RemoteHost:        "127.0.0.1",
		RemotePort:        3306,
		LocalAddr:         "127.0.0.1",
		LocalPort:         13306,
		ConnectTimeoutSec: 10,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	payload := ProfileTransferFile{
		SchemaVersion: profileTransferSchemaVersion,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		Profiles: []ProfileTransferProfile{
			{Title: "数据库隧道", SSHHost: "ssh2.example.com", SSHPort: 22, SSHUser: "root", SSHPassword: "x", RemoteHost: "127.0.0.1", RemotePort: 3307, LocalAddr: "127.0.0.1", LocalPort: 13307, ConnectTimeoutSec: 10},
			{Title: "缓存隧道", SSHHost: "ssh3.example.com", SSHPort: 22, SSHUser: "root", SSHPassword: "y", RemoteHost: "127.0.0.1", RemotePort: 6379, LocalAddr: "127.0.0.1", LocalPort: 16379, ConnectTimeoutSec: 10},
			{Title: "", SSHHost: "ssh4.example.com", SSHPort: 22, SSHUser: "root", SSHPassword: "z", RemoteHost: "127.0.0.1", RemotePort: 5432, LocalAddr: "127.0.0.1", LocalPort: 15432, ConnectTimeoutSec: 10},
			{Title: "缓存隧道", SSHHost: "ssh5.example.com", SSHPort: 22, SSHUser: "root", SSHPassword: "w", RemoteHost: "127.0.0.1", RemotePort: 6380, LocalAddr: "127.0.0.1", LocalPort: 16380, ConnectTimeoutSec: 10},
		},
	}
	data, _ := json.Marshal(payload)
	preview, err := store.PreviewImport(data)
	if err != nil {
		t.Fatalf("PreviewImport: %v", err)
	}
	if len(preview.Conflicts) != 1 || preview.Conflicts[0].Title != "数据库隧道" {
		t.Fatalf("unexpected conflicts: %+v", preview.Conflicts)
	}
	if len(preview.Added) != 1 || preview.Added[0].Title != "缓存隧道" {
		t.Fatalf("unexpected added: %+v", preview.Added)
	}
	if len(preview.Invalid) != 2 {
		t.Fatalf("invalid count = %d, want 2", len(preview.Invalid))
	}
}

func TestProfileStoreApplyImportMergeAndSkip(t *testing.T) {
	store, err := NewProfileStore(filepath.Join(t.TempDir(), "profiles.json"))
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}
	existing, err := store.Save(Profile{
		Title:             "数据库隧道",
		SSHHost:           "ssh.example.com",
		SSHPort:           22,
		SSHUser:           "root",
		SSHKeyPath:        "~/.ssh/id_rsa",
		RemoteHost:        "127.0.0.1",
		RemotePort:        3306,
		LocalAddr:         "127.0.0.1",
		LocalPort:         13306,
		ConnectTimeoutSec: 10,
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	payload := ProfileTransferFile{
		SchemaVersion: profileTransferSchemaVersion,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		Profiles: []ProfileTransferProfile{
			{Title: "数据库隧道", SSHHost: "ssh2.example.com", SSHPort: 2222, SSHUser: "admin", SSHPassword: "override", RemoteHost: "10.0.0.1", RemotePort: 3307, LocalAddr: "127.0.0.1", LocalPort: 23307, ConnectTimeoutSec: 12},
			{Title: "缓存隧道", SSHHost: "ssh3.example.com", SSHPort: 22, SSHUser: "root", SSHPassword: "new", RemoteHost: "127.0.0.1", RemotePort: 6379, LocalAddr: "127.0.0.1", LocalPort: 16379, ConnectTimeoutSec: 10},
		},
	}
	data, _ := json.Marshal(payload)

	skipResult, err := store.ApplyImport(data, "skip")
	if err != nil {
		t.Fatalf("ApplyImport(skip): %v", err)
	}
	if skipResult.Added != 1 || skipResult.Skipped != 1 || skipResult.Merged != 0 {
		t.Fatalf("unexpected skip result: %+v", skipResult)
	}
	skipped, ok := store.findByTitle("数据库隧道")
	if !ok || skipped.ID != existing.ID || skipped.SSHHost != "ssh.example.com" {
		t.Fatalf("existing profile changed unexpectedly: %+v", skipped)
	}

	store2, err := NewProfileStore(filepath.Join(t.TempDir(), "profiles-merge.json"))
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}
	existing2, _ := store2.Save(existing)
	mergeResult, err := store2.ApplyImport(data, "merge")
	if err != nil {
		t.Fatalf("ApplyImport(merge): %v", err)
	}
	if mergeResult.Added != 1 || mergeResult.Merged != 1 || mergeResult.Skipped != 0 {
		t.Fatalf("unexpected merge result: %+v", mergeResult)
	}
	merged, ok := store2.findByTitle("数据库隧道")
	if !ok {
		t.Fatal("merged profile not found")
	}
	if merged.ID != existing2.ID {
		t.Fatalf("merged profile ID changed: got %q want %q", merged.ID, existing2.ID)
	}
	if merged.SSHHost != "ssh2.example.com" || merged.LocalPort != 23307 {
		t.Fatalf("merged profile not updated: %+v", merged)
	}
}
