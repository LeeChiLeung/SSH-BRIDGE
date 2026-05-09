package main

import (
	"io"
	"log"
	"strings"
	"testing"
)

func TestStartAutoStartProfilesStartsEnabledProfiles(t *testing.T) {
	t.Parallel()

	sshAddr, closeSSH := startTestSSHServer(t, "secret")
	defer closeSSH()

	echoAddr, closeEcho := startEchoServer(t, "auto")
	defer closeEcho()

	store := newTestProfileStore(t)
	autoCfg := testTunnelConfig(sshAddr, echoAddr, "secret")
	autoCfg.Title = "auto"
	autoProfile, err := store.Save(profileFromConfig("auto", autoCfg))
	if err != nil {
		t.Fatalf("save auto profile: %v", err)
	}
	autoProfile.AutoStart = true
	if _, err := store.Save(autoProfile); err != nil {
		t.Fatalf("update auto profile: %v", err)
	}

	manualCfg := testTunnelConfig(sshAddr, echoAddr, "secret")
	manualCfg.Title = "manual"
	manualProfile, err := store.Save(profileFromConfig("manual", manualCfg))
	if err != nil {
		t.Fatalf("save manual profile: %v", err)
	}

	controller := NewTunnelController(log.New(io.Discard, "", 0), store)
	defer controller.StopAll()

	startAutoStartProfiles(controller, store, log.New(io.Discard, "", 0))

	waitForState(t, controller, autoProfile.ID, TunnelStateConnected)
	if got := controller.Snapshot(manualProfile.ID).State; got != TunnelStateDisconnected {
		t.Fatalf("manual profile state = %s, want disconnected", got)
	}
	assertForwarding(t, controller.Snapshot(autoProfile.ID).Endpoint, "auto")
}

func TestStartAutoStartProfilesIsolatesFailures(t *testing.T) {
	t.Parallel()

	sshAddr, closeSSH := startTestSSHServer(t, "secret")
	defer closeSSH()

	echoAddr, closeEcho := startEchoServer(t, "steady")
	defer closeEcho()

	store := newTestProfileStore(t)
	goodCfg := testTunnelConfig(sshAddr, echoAddr, "secret")
	goodCfg.Title = "good"
	goodProfile, err := store.Save(profileFromConfig("good", goodCfg))
	if err != nil {
		t.Fatalf("save good profile: %v", err)
	}
	goodProfile.AutoStart = true
	if _, err := store.Save(goodProfile); err != nil {
		t.Fatalf("update good profile: %v", err)
	}

	badCfg := testTunnelConfig(sshAddr, echoAddr, "secret")
	badCfg.Title = "bad"
	badCfg.LocalPort = unusedPort(t)
	blocker, err := prepareListener(cliOptions{TunnelConfig: badCfg})
	if err != nil {
		t.Fatalf("prepare blocker listener: %v", err)
	}
	defer blocker.Close()

	badProfile, err := store.Save(profileFromConfig("bad", badCfg))
	if err != nil {
		t.Fatalf("save bad profile: %v", err)
	}
	badProfile.AutoStart = true
	if _, err := store.Save(badProfile); err != nil {
		t.Fatalf("update bad profile: %v", err)
	}

	logBuf := &strings.Builder{}
	logger := log.New(logBuf, "", 0)
	controller := NewTunnelController(logger, store)
	defer controller.StopAll()

	startAutoStartProfiles(controller, store, logger)

	waitForState(t, controller, goodProfile.ID, TunnelStateConnected)
	waitForState(t, controller, badProfile.ID, TunnelStateDisconnected)

	if got := controller.Snapshot(badProfile.ID).LastError; got == "" {
		t.Fatal("bad profile last error = empty, want bind failure")
	}
	if !strings.Contains(logBuf.String(), `autostart profile "bad"`) {
		t.Fatalf("autostart failure log missing, got: %s", logBuf.String())
	}
	assertForwarding(t, controller.Snapshot(goodProfile.ID).Endpoint, "steady")
}
