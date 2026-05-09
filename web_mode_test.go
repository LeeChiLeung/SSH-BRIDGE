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

func TestNextWebAddrUsesAdjacentPort(t *testing.T) {
	got, err := nextWebAddr("127.0.0.1:8848")
	if err != nil {
		t.Fatalf("nextWebAddr: %v", err)
	}
	if got != "127.0.0.1:8849" {
		t.Fatalf("nextWebAddr = %q, want %q", got, "127.0.0.1:8849")
	}
}
