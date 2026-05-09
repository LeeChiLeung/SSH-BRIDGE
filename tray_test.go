package main

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
	"testing"
	"time"
)

func TestTrayAppMenuAndOpenAction(t *testing.T) {
	t.Parallel()

	manager := newFakeTrayManager()
	opener := &fakeBrowserOpener{}
	app := &trayApp{
		manager:   manager,
		opener:    opener,
		logger:    log.New(io.Discard, "", 0),
		newWebURL: "http://127.0.0.1:8849",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	manager.waitReady(t)
	if got := manager.menuLabels(); len(got) != 2 || got[0] != trayOpenNewConsole || got[1] != trayQuitLabel {
		t.Fatalf("unexpected tray menu labels: %v", got)
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("tray app returned error: %v", err)
	}
}

func TestTrayAppNewConsoleAction(t *testing.T) {
	t.Parallel()

	manager := newFakeTrayManager()
	opener := &fakeBrowserOpener{}
	app := &trayApp{
		manager:   manager,
		opener:    opener,
		logger:    log.New(io.Discard, "", 0),
		newWebURL: "http://127.0.0.1:8849",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	manager.waitReady(t)
	manager.click(trayOpenNewConsole)
	waitForCondition(t, func() bool {
		return opener.lastURL() == "http://127.0.0.1:8849"
	})

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("tray app returned error: %v", err)
	}
}

func TestTrayAppQuitActionTriggersShutdown(t *testing.T) {
	t.Parallel()

	manager := newFakeTrayManager()
	quitCalled := make(chan struct{}, 1)
	app := &trayApp{
		manager:   manager,
		opener:    &fakeBrowserOpener{},
		logger:    log.New(io.Discard, "", 0),
		newWebURL: "http://127.0.0.1:8849",
		onQuit: func() {
			quitCalled <- struct{}{}
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	manager.waitReady(t)
	manager.click(trayQuitLabel)
	waitForCondition(t, func() bool {
		select {
		case <-quitCalled:
			return true
		default:
			return false
		}
	})

	if err := <-done; err != nil {
		t.Fatalf("tray app returned error: %v", err)
	}
}

func TestTrayAppOpenFailureKeepsTrayAlive(t *testing.T) {
	t.Parallel()

	manager := newFakeTrayManager()
	app := &trayApp{
		manager:   manager,
		opener:    failingBrowserOpener{},
		logger:    log.New(io.Discard, "", 0),
		newWebURL: "http://127.0.0.1:8849",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	manager.waitReady(t)
	manager.click(trayOpenNewConsole)

	select {
	case err := <-done:
		t.Fatalf("tray app exited unexpectedly after browser open failure: %v", err)
	case <-time.After(150 * time.Millisecond):
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("tray app returned error: %v", err)
	}
}

func TestResolveTrayIconFallsBackWhenFileMissing(t *testing.T) {
	t.Parallel()

	iconData, err := loadTrayIcon("icon/32x32.ico")
	if err != nil {
		t.Fatalf("load tray icon from repo asset: %v", err)
	}
	if len(iconData) == 0 {
		t.Fatal("expected decoded tray icon bytes")
	}

	fallback := embeddedTrayIcon()
	if len(fallback) == 0 {
		t.Fatal("expected embedded fallback icon bytes")
	}
}

func TestTrayIconCandidatesForWindowsPreferICO(t *testing.T) {
	t.Parallel()

	paths := trayIconCandidatesForOS("windows")
	if len(paths) == 0 {
		t.Fatal("expected windows tray icon candidates")
	}
	if paths[0] != "icon/32x32.ico" {
		t.Fatalf("first windows icon candidate = %q, want %q", paths[0], "icon/32x32.ico")
	}
}

func TestResolveTrayIconForWindowsUsesEmbeddedICO(t *testing.T) {
	t.Parallel()

	iconData := resolveTrayIconForOS("windows")
	if len(iconData) == 0 {
		t.Fatal("expected windows tray icon data")
	}
}

type fakeTrayManager struct {
	mu        sync.Mutex
	ready     chan struct{}
	exit      chan struct{}
	menuItems map[string]*fakeTrayMenuItem
	labels    []string
}

type fakeTrayMenuItem struct {
	onClick func()
}

func newFakeTrayManager() *fakeTrayManager {
	return &fakeTrayManager{
		ready:     make(chan struct{}),
		exit:      make(chan struct{}),
		menuItems: make(map[string]*fakeTrayMenuItem),
	}
}

func (m *fakeTrayManager) Run(onReady func(), onExit func()) {
	onReady()
	close(m.ready)
	<-m.exit
	onExit()
}

func (m *fakeTrayManager) Quit() {
	select {
	case <-m.exit:
	default:
		close(m.exit)
	}
}

func (m *fakeTrayManager) SetTitle(string)   {}
func (m *fakeTrayManager) SetTooltip(string) {}
func (m *fakeTrayManager) SetIcon([]byte)    {}

func (m *fakeTrayManager) AddMenuItem(title string, tooltip string) trayMenuItem {
	m.mu.Lock()
	defer m.mu.Unlock()
	item := &fakeTrayMenuItem{}
	m.menuItems[title] = item
	m.labels = append(m.labels, title)
	return item
}

func (m *fakeTrayManager) waitReady(t *testing.T) {
	t.Helper()
	select {
	case <-m.ready:
	case <-time.After(2 * time.Second):
		t.Fatal("tray manager did not become ready")
	}
}

func (m *fakeTrayManager) click(label string) {
	m.mu.Lock()
	item := m.menuItems[label]
	m.mu.Unlock()
	if item != nil && item.onClick != nil {
		item.onClick()
	}
}

func (m *fakeTrayManager) menuLabels() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.labels))
	copy(out, m.labels)
	return out
}

func (i *fakeTrayMenuItem) OnClick(fn func()) {
	i.onClick = fn
}

type fakeBrowserOpener struct {
	mu  sync.Mutex
	url string
}

func (o *fakeBrowserOpener) Open(url string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.url = url
	return nil
}

func (o *fakeBrowserOpener) lastURL() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.url
}

type failingBrowserOpener struct{}

func (failingBrowserOpener) Open(string) error {
	return errors.New("open failed")
}

func waitForCondition(t *testing.T, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition not satisfied before timeout")
}
