package main

import (
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileFromRequestCapturesPassword(t *testing.T) {
	form := url.Values{
		"id":                   {"profile-1"},
		"title":                {"测试隧道"},
		"ssh_host":             {"ssh.example.com"},
		"ssh_port":             {"22"},
		"ssh_user":             {"user"},
		"ssh_password":         {"secret"},
		"ssh_key_path":         {"~/.ssh/example_id_rsa"},
		"remote_host":          {"127.0.0.1"},
		"remote_port":          {"5432"},
		"local_addr":           {"127.0.0.1"},
		"local_port":           {"15432"},
		"connect_timeout_sec":  {"10"},
		"known_hosts_path":     {"/tmp/known_hosts"},
		"auto_start":           {"on"},
		"insecure_ignore_host": {"on"},
	}

	req := httptest.NewRequest("POST", "/profiles/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	profile, err := profileFromRequest(req)
	if err != nil {
		t.Fatalf("profileFromRequest: %v", err)
	}
	if profile.SSHPassword != "secret" {
		t.Fatalf("SSHPassword = %q, want %q", profile.SSHPassword, "secret")
	}
	if profile.SSHKeyPath != "~/.ssh/example_id_rsa" {
		t.Fatalf("SSHKeyPath = %q", profile.SSHKeyPath)
	}
	if !profile.AutoStart {
		t.Fatal("AutoStart = false, want true")
	}
}

func TestProfileStorePersistsPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profiles.json")
	store, err := NewProfileStore(path)
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}

	saved, err := store.Save(Profile{
		ID:                 "profile-1",
		Title:              "测试隧道",
		SSHHost:            "ssh.example.com",
		SSHPort:            22,
		SSHUser:            "user",
		SSHPassword:        "secret",
		SSHKeyPath:         "~/.ssh/example_id_rsa",
		RemoteHost:         "127.0.0.1",
		RemotePort:         5432,
		LocalAddr:          "127.0.0.1",
		LocalPort:          15432,
		ConnectTimeoutSec:  10,
		KnownHostsPath:     "/tmp/known_hosts",
		AutoStart:          true,
		InsecureIgnoreHost: true,
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := NewProfileStore(path)
	if err != nil {
		t.Fatalf("reload store: %v", err)
	}
	profile, ok := reloaded.Get(saved.ID)
	if !ok {
		t.Fatalf("reloaded profile %q not found", saved.ID)
	}
	if profile.SSHPassword != "secret" {
		t.Fatalf("reloaded SSHPassword = %q, want %q", profile.SSHPassword, "secret")
	}
	if profile.TunnelConfig().SSHPassword != "secret" {
		t.Fatalf("tunnel config SSHPassword = %q, want %q", profile.TunnelConfig().SSHPassword, "secret")
	}
	if !profile.AutoStart {
		t.Fatal("reloaded AutoStart = false, want true")
	}
}
