package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"os"
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

func TestProfileFromRequestUploadKeyFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("id", "profile-upload")
	_ = writer.WriteField("title", "上传私钥测试")
	_ = writer.WriteField("ssh_host", "ssh.example.com")
	_ = writer.WriteField("ssh_port", "22")
	_ = writer.WriteField("ssh_user", "ec2-user")
	_ = writer.WriteField("remote_host", "127.0.0.1")
	_ = writer.WriteField("remote_port", "5432")
	_ = writer.WriteField("local_addr", "127.0.0.1")
	_ = writer.WriteField("local_port", "15432")
	_ = writer.WriteField("connect_timeout_sec", "10")
	_ = writer.WriteField("ssh_key_path", "/old/path/should/be/overridden")

	part, err := writer.CreateFormFile("ssh_key_file", "my-ec2-key.pem")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(pemBytes); err != nil {
		t.Fatalf("write pem part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}

	req := httptest.NewRequest("POST", "/profiles/save", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	profile, err := profileFromRequest(req)
	if err != nil {
		t.Fatalf("profileFromRequest: %v", err)
	}

	if profile.SSHKeyPath == "/old/path/should/be/overridden" || profile.SSHKeyPath == "" {
		t.Fatalf("SSHKeyPath = %q, expected newly saved path", profile.SSHKeyPath)
	}
	if !strings.HasSuffix(profile.SSHKeyPath, "my-ec2-key.pem") {
		t.Fatalf("SSHKeyPath = %q, expected filename suffix my-ec2-key.pem", profile.SSHKeyPath)
	}

	content, err := os.ReadFile(profile.SSHKeyPath)
	if err != nil {
		t.Fatalf("read saved key file: %v", err)
	}
	if !bytes.Equal(content, pemBytes) {
		t.Fatalf("saved key file content does not match uploaded pem bytes")
	}

	info, err := os.Stat(profile.SSHKeyPath)
	if err != nil {
		t.Fatalf("stat saved key file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("key file perm = %v, want 0600", perm)
	}
}

func TestProfileFromRequestUploadInvalidKeyFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("id", "profile-bad-upload")
	_ = writer.WriteField("title", "非法私钥测试")
	_ = writer.WriteField("ssh_host", "ssh.example.com")
	_ = writer.WriteField("ssh_port", "22")
	_ = writer.WriteField("ssh_user", "ec2-user")
	_ = writer.WriteField("remote_host", "127.0.0.1")
	_ = writer.WriteField("remote_port", "5432")
	_ = writer.WriteField("local_addr", "127.0.0.1")
	_ = writer.WriteField("local_port", "15432")
	_ = writer.WriteField("connect_timeout_sec", "10")

	part, err := writer.CreateFormFile("ssh_key_file", "bad.pem")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte("not a valid pem private key")); err != nil {
		t.Fatalf("write bad part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}

	req := httptest.NewRequest("POST", "/profiles/save", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, err = profileFromRequest(req)
	if err == nil {
		t.Fatal("expected error on invalid key upload, got nil")
	}
	if !strings.Contains(err.Error(), "invalid SSH private key") {
		t.Fatalf("error = %q, expected 'invalid SSH private key'", err.Error())
	}
}

