package main

import (
	"io"
	"log"
	"testing"
)

func TestManagedWebModesSelectDifferentImplementations(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	opts := cliOptions{
		TunnelConfig: defaultTunnelConfig(),
		WebAddr:      "127.0.0.1:18848",
	}
	logger := log.New(io.Discard, "", 0)

	legacy, err := newManagedWebMode(opts, logger)
	if err != nil {
		t.Fatalf("newManagedWebMode: %v", err)
	}
	if _, ok := legacy.server.(*WebServer); !ok {
		t.Fatalf("legacy server type = %T, want *WebServer", legacy.server)
	}

	modern, err := newManagedNewWebMode(opts, logger)
	if err != nil {
		t.Fatalf("newManagedNewWebMode: %v", err)
	}
	if _, ok := modern.server.(*TemplWebServer); !ok {
		t.Fatalf("new server type = %T, want *TemplWebServer", modern.server)
	}
}

func TestManagedRuntimeCanBeSharedAcrossWebModes(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	opts := cliOptions{
		TunnelConfig: defaultTunnelConfig(),
		WebAddr:      "127.0.0.1:18848",
	}
	logger := log.New(io.Discard, "", 0)

	runtime, err := newManagedRuntime(opts, logger)
	if err != nil {
		t.Fatalf("newManagedRuntime: %v", err)
	}
	if runtime.store == nil || runtime.controller == nil {
		t.Fatal("managed runtime is incomplete")
	}

	legacy, err := NewWebServer(opts.WebAddr, runtime.controller, runtime.store, logger)
	if err != nil {
		t.Fatalf("NewWebServer: %v", err)
	}
	modern := NewTemplWebServer("127.0.0.1:18849", newUIRouter(runtime.controller, runtime.store, ""))

	if legacy.store != runtime.store || legacy.controller != runtime.controller {
		t.Fatal("legacy web server does not share managed runtime")
	}
	if modern == nil {
		t.Fatal("modern web server is nil")
	}
}

func TestNextWebAddrUsesAdjacentPort(t *testing.T) {
	got, err := nextWebAddr("127.0.0.1:8848")
	if err != nil {
		t.Fatalf("nextWebAddr: %v", err)
	}
	if got != "127.0.0.1:8849" {
		t.Fatalf("nextWebAddr = %q, want %q", got, "127.0.0.1:8849")
	}
}

func TestShouldDefaultToTrayForWindows(t *testing.T) {
	opts := cliOptions{WebAddr: "127.0.0.1:8848"}
	if !shouldDefaultToTrayForOS("windows", opts) {
		t.Fatal("expected windows no-arg launch to default to tray mode")
	}
	if shouldDefaultToTrayForOS("linux", opts) {
		t.Fatal("did not expect linux launch to default to tray mode")
	}
}

func TestShouldDefaultToTraySkipsExplicitModes(t *testing.T) {
	opts := cliOptions{TrayMode: true, WebAddr: "127.0.0.1:8848"}
	if shouldDefaultToTrayForOS("windows", opts) {
		t.Fatal("did not expect explicit tray mode to trigger implicit tray default")
	}

	opts = cliOptions{WebNewMode: true, WebAddr: "127.0.0.1:8848"}
	if shouldDefaultToTrayForOS("windows", opts) {
		t.Fatal("did not expect explicit web-new mode to trigger implicit tray default")
	}
}
