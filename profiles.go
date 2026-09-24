package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Profile struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	SSHHost            string     `json:"sshHost"`
	SSHPort            int        `json:"sshPort"`
	SSHUser            string     `json:"sshUser"`
	SSHPassword        string     `json:"sshPassword"`
	SSHKeyPath         string     `json:"sshKeyPath"`
	RemoteHost         string     `json:"remoteHost"`
	RemotePort         int        `json:"remotePort"`
	LocalAddr          string     `json:"localAddr"`
	LocalPort          int        `json:"localPort"`
	ConnectTimeoutSec  int        `json:"connectTimeoutSec"`
	KnownHostsPath     string     `json:"knownHostsPath"`
	InsecureIgnoreHost bool       `json:"insecureIgnoreHost"`
	AutoStart          bool       `json:"autoStart"`
	LastEnabledAt      *time.Time `json:"lastEnabledAt,omitempty"`
}

func (p Profile) TunnelConfig() TunnelConfig {
	timeout := time.Duration(p.ConnectTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return TunnelConfig{
		Title:              p.Title,
		SSHHost:            p.SSHHost,
		SSHPort:            p.SSHPort,
		SSHUser:            p.SSHUser,
		SSHPassword:        p.SSHPassword,
		SSHKeyPath:         p.SSHKeyPath,
		RemoteHost:         p.RemoteHost,
		RemotePort:         p.RemotePort,
		LocalAddr:          p.LocalAddr,
		LocalPort:          p.LocalPort,
		ConnectTimeout:     timeout,
		KnownHostsPath:     p.KnownHostsPath,
		InsecureIgnoreHost: p.InsecureIgnoreHost,
	}
}

func profileFromConfig(id string, cfg TunnelConfig) Profile {
	return Profile{
		ID:                 id,
		Title:              cfg.DisplayTitle(),
		SSHHost:            cfg.SSHHost,
		SSHPort:            cfg.SSHPort,
		SSHUser:            cfg.SSHUser,
		SSHPassword:        cfg.SSHPassword,
		SSHKeyPath:         cfg.SSHKeyPath,
		RemoteHost:         cfg.RemoteHost,
		RemotePort:         cfg.RemotePort,
		LocalAddr:          cfg.LocalAddr,
		LocalPort:          cfg.LocalPort,
		ConnectTimeoutSec:  int(cfg.ConnectTimeout.Seconds()),
		KnownHostsPath:     cfg.KnownHostsPath,
		InsecureIgnoreHost: cfg.InsecureIgnoreHost,
	}
}

type ProfileStore struct {
	mu       sync.Mutex
	path     string
	profiles []Profile
}

func NewProfileStore(path string) (*ProfileStore, error) {
	store := &ProfileStore{path: path, profiles: []Profile{}}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *ProfileStore) EnsureDefault(cfg TunnelConfig) (Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.profiles) > 0 {
		return s.profiles[0], nil
	}
	profile := profileFromConfig(newProfileID(), cfg)
	s.profiles = append(s.profiles, profile)
	return profile, s.saveLocked()
}

func (s *ProfileStore) List() []Profile {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Profile, len(s.profiles))
	copy(out, s.profiles)
	return out
}

func (s *ProfileStore) Get(id string) (Profile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, profile := range s.profiles {
		if profile.ID == id {
			return profile, true
		}
	}
	return Profile{}, false
}

func (s *ProfileStore) Save(profile Profile) (Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(profile.ID) == "" {
		profile.ID = newProfileID()
	}
	if strings.TrimSpace(profile.Title) == "" {
		profile.Title = defaultTitle
	}
	for idx, existing := range s.profiles {
		if existing.ID == profile.ID {
			profile.LastEnabledAt = existing.LastEnabledAt
			s.profiles[idx] = profile
			return profile, s.saveLocked()
		}
	}
	s.profiles = append(s.profiles, profile)
	return profile, s.saveLocked()
}

func (s *ProfileStore) Duplicate(id string) (Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.profiles {
		if existing.ID == id {
			clone := existing
			clone.ID = newProfileID()
			clone.Title = existing.Title + " 副本"
			clone.LastEnabledAt = nil
			s.profiles = append(s.profiles, clone)
			return clone, s.saveLocked()
		}
	}
	return Profile{}, fmt.Errorf("profile not found")
}

func (s *ProfileStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for idx, existing := range s.profiles {
		if existing.ID == id {
			s.profiles = append(s.profiles[:idx], s.profiles[idx+1:]...)
			return s.saveLocked()
		}
	}
	return fmt.Errorf("profile not found")
}

func (s *ProfileStore) TouchLastEnabled(id string, when time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for idx := range s.profiles {
		if s.profiles[idx].ID == id {
			ts := when
			s.profiles[idx].LastEnabledAt = &ts
			return s.saveLocked()
		}
	}
	return nil
}

func (s *ProfileStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var payload struct {
		Profiles []Profile `json:"profiles"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	s.profiles = payload.Profiles
	return nil
}

func (s *ProfileStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	payload := struct {
		Profiles []Profile `json:"profiles"`
	}{Profiles: s.profiles}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func defaultProfileStorePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(".", "profiles.json")
	}
	return filepath.Join(configDir, "timber-ssh-go", "profiles.json")
}

func newProfileID() string {
	return fmt.Sprintf("profile-%d", time.Now().UnixNano())
}

func expandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if strings.HasPrefix(path, "~/") {
		home, err := osUserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}
	return path, nil
}

func defaultKeysDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(".", "keys")
	}
	return filepath.Join(configDir, "timber-ssh-go", "keys")
}

func saveUploadedPrivateKey(filename string, data []byte) (string, error) {
	if _, err := ssh.ParsePrivateKey(data); err != nil {
		return "", fmt.Errorf("invalid SSH private key: %w", err)
	}
	keysDir := defaultKeysDir()
	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		return "", fmt.Errorf("create keys directory: %w", err)
	}
	cleanName := filepath.Base(filename)
	if cleanName == "." || cleanName == "/" || cleanName == "" {
		cleanName = "id_rsa.pem"
	}
	targetName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), cleanName)
	targetPath := filepath.Join(keysDir, targetName)
	if err := os.WriteFile(targetPath, data, 0o600); err != nil {
		return "", fmt.Errorf("save private key: %w", err)
	}
	return targetPath, nil
}

