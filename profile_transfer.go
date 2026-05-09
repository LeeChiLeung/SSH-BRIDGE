package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const profileTransferSchemaVersion = "timber-ssh-go/profile-export@v1"

type ProfileTransferFile struct {
	SchemaVersion string                   `json:"schemaVersion"`
	ExportedAt    string                   `json:"exportedAt"`
	Profiles      []ProfileTransferProfile `json:"profiles"`
}

type ProfileTransferProfile struct {
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

type ProfileImportInvalid struct {
	Index  int    `json:"index"`
	Title  string `json:"title"`
	Reason string `json:"reason"`
}

type ProfileImportConflict struct {
	Title    string                 `json:"title"`
	Existing ProfileTransferProfile `json:"existing"`
	Incoming ProfileTransferProfile `json:"incoming"`
}

type ProfileImportPreview struct {
	SchemaVersion string                   `json:"schemaVersion"`
	Added         []ProfileTransferProfile `json:"added"`
	Conflicts     []ProfileImportConflict  `json:"conflicts"`
	Invalid       []ProfileImportInvalid   `json:"invalid"`
}

type ProfileImportResult struct {
	Added   int `json:"added"`
	Merged  int `json:"merged"`
	Skipped int `json:"skipped"`
	Invalid int `json:"invalid"`
}

func encodeImportPayload(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func decodeImportPayload(encoded string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode import payload: %w", err)
	}
	return data, nil
}

func transferProfileFromProfile(profile Profile) ProfileTransferProfile {
	return ProfileTransferProfile{
		Title:              profile.Title,
		SSHHost:            profile.SSHHost,
		SSHPort:            profile.SSHPort,
		SSHUser:            profile.SSHUser,
		SSHPassword:        profile.SSHPassword,
		SSHKeyPath:         profile.SSHKeyPath,
		RemoteHost:         profile.RemoteHost,
		RemotePort:         profile.RemotePort,
		LocalAddr:          profile.LocalAddr,
		LocalPort:          profile.LocalPort,
		ConnectTimeoutSec:  profile.ConnectTimeoutSec,
		KnownHostsPath:     profile.KnownHostsPath,
		InsecureIgnoreHost: profile.InsecureIgnoreHost,
		AutoStart:          profile.AutoStart,
		LastEnabledAt:      profile.LastEnabledAt,
	}
}

func profileFromTransferProfile(item ProfileTransferProfile) Profile {
	return Profile{
		Title:              strings.TrimSpace(item.Title),
		SSHHost:            strings.TrimSpace(item.SSHHost),
		SSHPort:            item.SSHPort,
		SSHUser:            strings.TrimSpace(item.SSHUser),
		SSHPassword:        item.SSHPassword,
		SSHKeyPath:         strings.TrimSpace(item.SSHKeyPath),
		RemoteHost:         strings.TrimSpace(item.RemoteHost),
		RemotePort:         item.RemotePort,
		LocalAddr:          strings.TrimSpace(item.LocalAddr),
		LocalPort:          item.LocalPort,
		ConnectTimeoutSec:  item.ConnectTimeoutSec,
		KnownHostsPath:     strings.TrimSpace(item.KnownHostsPath),
		InsecureIgnoreHost: item.InsecureIgnoreHost,
		AutoStart:          item.AutoStart,
		LastEnabledAt:      item.LastEnabledAt,
	}
}

func (s *ProfileStore) ExportTransferFile(now time.Time) ProfileTransferFile {
	profiles := s.List()
	items := make([]ProfileTransferProfile, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, transferProfileFromProfile(profile))
	}
	return ProfileTransferFile{
		SchemaVersion: profileTransferSchemaVersion,
		ExportedAt:    now.UTC().Format(time.RFC3339),
		Profiles:      items,
	}
}

func (s *ProfileStore) ExportTransferJSON(now time.Time) ([]byte, error) {
	payload := s.ExportTransferFile(now)
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal export payload: %w", err)
	}
	return data, nil
}

func (s *ProfileStore) PreviewImport(data []byte) (ProfileImportPreview, error) {
	payload, err := parseTransferFile(data)
	if err != nil {
		return ProfileImportPreview{}, err
	}

	existing := s.List()
	byTitle := make(map[string]Profile, len(existing))
	for _, profile := range existing {
		byTitle[strings.TrimSpace(profile.Title)] = profile
	}

	seenImported := make(map[string]int)
	preview := ProfileImportPreview{SchemaVersion: payload.SchemaVersion}
	for idx, item := range payload.Profiles {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			preview.Invalid = append(preview.Invalid, ProfileImportInvalid{Index: idx, Title: title, Reason: "title is required"})
			continue
		}
		if firstIdx, ok := seenImported[title]; ok {
			preview.Invalid = append(preview.Invalid, ProfileImportInvalid{Index: idx, Title: title, Reason: fmt.Sprintf("duplicate title in import file (first seen at index %d)", firstIdx)})
			continue
		}
		seenImported[title] = idx
		if existingProfile, ok := byTitle[title]; ok {
			preview.Conflicts = append(preview.Conflicts, ProfileImportConflict{
				Title:    title,
				Existing: transferProfileFromProfile(existingProfile),
				Incoming: item,
			})
			continue
		}
		preview.Added = append(preview.Added, item)
	}
	return preview, nil
}

func (s *ProfileStore) ApplyImport(data []byte, policy string) (ProfileImportResult, error) {
	preview, err := s.PreviewImport(data)
	if err != nil {
		return ProfileImportResult{}, err
	}
	switch policy {
	case "skip", "merge":
	default:
		return ProfileImportResult{}, fmt.Errorf("unsupported conflict policy: %s", policy)
	}

	result := ProfileImportResult{
		Invalid: len(preview.Invalid),
	}

	for _, item := range preview.Added {
		if _, err := s.Save(profileFromTransferProfile(item)); err != nil {
			return result, err
		}
		result.Added++
	}

	if policy == "skip" {
		result.Skipped = len(preview.Conflicts)
		return result, nil
	}

	for _, conflict := range preview.Conflicts {
		existing, ok := s.findByTitle(conflict.Title)
		if !ok {
			result.Invalid++
			continue
		}
		merged := profileFromTransferProfile(conflict.Incoming)
		merged.ID = existing.ID
		merged.Title = existing.Title
		merged.LastEnabledAt = existing.LastEnabledAt
		if _, err := s.Save(merged); err != nil {
			return result, err
		}
		result.Merged++
	}

	return result, nil
}

func (s *ProfileStore) findByTitle(title string) (Profile, bool) {
	title = strings.TrimSpace(title)
	for _, profile := range s.List() {
		if strings.TrimSpace(profile.Title) == title {
			return profile, true
		}
	}
	return Profile{}, false
}

func parseTransferFile(data []byte) (ProfileTransferFile, error) {
	var payload ProfileTransferFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return ProfileTransferFile{}, fmt.Errorf("parse import file: %w", err)
	}
	if strings.TrimSpace(payload.SchemaVersion) != profileTransferSchemaVersion {
		return ProfileTransferFile{}, fmt.Errorf("unsupported schema version: %s", payload.SchemaVersion)
	}
	if payload.Profiles == nil {
		return ProfileTransferFile{}, fmt.Errorf("profiles array is required")
	}
	return payload, nil
}
