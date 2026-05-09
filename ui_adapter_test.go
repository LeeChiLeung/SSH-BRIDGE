package main

import (
	"io"
	"log"
	"testing"
)

func TestBuildUIProfileCardsUsesSharedSnapshotData(t *testing.T) {
	store := newTestProfileStore(t)
	profile, err := store.Save(Profile{
		ID:                "profile-1",
		Title:             "数据库隧道",
		SSHHost:           "ssh.example.com",
		SSHPort:           22,
		SSHUser:           "user",
		SSHKeyPath:        "~/.ssh/example_id_rsa",
		RemoteHost:        "127.0.0.1",
		RemotePort:        5432,
		LocalAddr:         "127.0.0.1",
		LocalPort:         15432,
		ConnectTimeoutSec: 10,
	})
	if err != nil {
		t.Fatalf("save profile: %v", err)
	}

	controller := NewTunnelController(log.New(io.Discard, "", 0), store)
	controller.mu.Lock()
	controller.snapshots[profile.ID] = &TunnelSnapshot{
		ProfileID: profile.ID,
		State:     TunnelStateConnected,
		Endpoint:  "127.0.0.1:23307",
	}
	controller.mu.Unlock()

	cards := buildUIProfileCards(controller, store, "zh-CN")
	if len(cards) != 1 {
		t.Fatalf("cards len = %d, want 1", len(cards))
	}
	if cards[0].LocalEndpoint != "127.0.0.1:23307" {
		t.Fatalf("LocalEndpoint = %q", cards[0].LocalEndpoint)
	}
	if cards[0].StateLabel != "已连接" {
		t.Fatalf("StateLabel = %q", cards[0].StateLabel)
	}
	if cards[0].StateTone != "green" {
		t.Fatalf("StateTone = %q", cards[0].StateTone)
	}
	if !cards[0].CanStop || cards[0].CanStart {
		t.Fatalf("card start/stop flags = start:%v stop:%v", cards[0].CanStart, cards[0].CanStop)
	}
}

func TestStateToneText(t *testing.T) {
	cases := []struct {
		name      string
		state     TunnelState
		lastError string
		want      string
	}{
		{name: "disconnected", state: TunnelStateDisconnected, want: "gray"},
		{name: "connecting", state: TunnelStateConnecting, want: "blue"},
		{name: "retrying", state: TunnelStateRetrying, want: "blue"},
		{name: "connected", state: TunnelStateConnected, want: "green"},
		{name: "failed", state: TunnelStateDisconnected, lastError: "auth failed", want: "red"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stateToneText(tc.state, tc.lastError); got != tc.want {
				t.Fatalf("stateToneText(%q, %q) = %q, want %q", tc.state, tc.lastError, got, tc.want)
			}
		})
	}
}
