package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type TunnelState string

const (
	TunnelStateDisconnected TunnelState = "disconnected"
	TunnelStateConnecting   TunnelState = "connecting"
	TunnelStateConnected    TunnelState = "connected"
	TunnelStateRetrying     TunnelState = "retrying"
)

type TunnelHealth struct {
	ListenerReady bool `json:"listenerReady"`
	SSHConnected  bool `json:"sshConnected"`
}

type LogEntry struct {
	At      time.Time `json:"at"`
	Message string    `json:"message"`
}

type TrafficStats struct {
	UploadBytesTotal    uint64 `json:"uploadBytesTotal"`
	DownloadBytesTotal  uint64 `json:"downloadBytesTotal"`
	UploadBytesPerSec   uint64 `json:"uploadBytesPerSec"`
	DownloadBytesPerSec uint64 `json:"downloadBytesPerSec"`
	ActiveConnections   int64  `json:"activeConnections"`
}

type TunnelSnapshot struct {
	ProfileID string       `json:"profileId"`
	State     TunnelState  `json:"state"`
	Endpoint  string       `json:"endpoint"`
	Health    TunnelHealth `json:"health"`
	LastError string       `json:"lastError"`
	Traffic   TrafficStats `json:"traffic"`
	Logs      []LogEntry   `json:"logs"`
}

type trafficMetrics struct {
	uploadTotal   atomic.Uint64
	downloadTotal atomic.Uint64
	uploadBPS     atomic.Uint64
	downloadBPS   atomic.Uint64
	active        atomic.Int64
}

func (m *trafficMetrics) AddUpload(n int) {
	if n > 0 {
		m.uploadTotal.Add(uint64(n))
	}
}

func (m *trafficMetrics) AddDownload(n int) {
	if n > 0 {
		m.downloadTotal.Add(uint64(n))
	}
}

func (m *trafficMetrics) IncActive() {
	m.active.Add(1)
}

func (m *trafficMetrics) DecActive() {
	m.active.Add(-1)
}

func (m *trafficMetrics) ResetRates() {
	m.uploadBPS.Store(0)
	m.downloadBPS.Store(0)
}

func (m *trafficMetrics) ResetLive() {
	m.ResetRates()
	m.active.Store(0)
}

func (m *trafficMetrics) Snapshot() TrafficStats {
	return TrafficStats{
		UploadBytesTotal:    m.uploadTotal.Load(),
		DownloadBytesTotal:  m.downloadTotal.Load(),
		UploadBytesPerSec:   m.uploadBPS.Load(),
		DownloadBytesPerSec: m.downloadBPS.Load(),
		ActiveConnections:   m.active.Load(),
	}
}

func (m *trafficMetrics) sampleLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	prevUpload := m.uploadTotal.Load()
	prevDownload := m.downloadTotal.Load()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentUpload := m.uploadTotal.Load()
			currentDownload := m.downloadTotal.Load()
			m.uploadBPS.Store(currentUpload - prevUpload)
			m.downloadBPS.Store(currentDownload - prevDownload)
			prevUpload = currentUpload
			prevDownload = currentDownload
		}
	}
}

type sessionStore struct {
	mu     sync.RWMutex
	client *ssh.Client
}

func (s *sessionStore) Get() *ssh.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.client
}

func (s *sessionStore) Set(client *ssh.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = client
}

func (s *sessionStore) Clear(target *ssh.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == target {
		s.client = nil
	}
}

type trackedTunnel struct {
	local  net.Conn
	remote net.Conn
	once   sync.Once
}

func (t *trackedTunnel) Close() {
	t.once.Do(func() {
		_ = t.local.Close()
		_ = t.remote.Close()
	})
}

type tunnelTracker struct {
	mu      sync.Mutex
	tunnels map[*trackedTunnel]struct{}
}

func newTunnelTracker() *tunnelTracker {
	return &tunnelTracker{tunnels: make(map[*trackedTunnel]struct{})}
}

func (t *tunnelTracker) Add(tunnel *trackedTunnel) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tunnels[tunnel] = struct{}{}
}

func (t *tunnelTracker) Remove(tunnel *trackedTunnel) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tunnels, tunnel)
}

func (t *tunnelTracker) CloseAll() {
	t.mu.Lock()
	tunnels := make([]*trackedTunnel, 0, len(t.tunnels))
	for tunnel := range t.tunnels {
		tunnels = append(tunnels, tunnel)
	}
	t.mu.Unlock()

	for _, tunnel := range tunnels {
		tunnel.Close()
	}
}

type tunnelRun struct {
	profileID string
	cfg       TunnelConfig
	ctx       context.Context
	cancel    context.CancelFunc
	done      chan struct{}
	listener  net.Listener
	tracker   *tunnelTracker
	sessions  *sessionStore
	traffic   *trafficMetrics
}

type TunnelController struct {
	mu           sync.RWMutex
	logger       *log.Logger
	profileStore *ProfileStore
	runs         map[string]*tunnelRun
	snapshots    map[string]*TunnelSnapshot
	maxLogs      int
}

func NewTunnelController(logger *log.Logger, store *ProfileStore) *TunnelController {
	return &TunnelController{
		logger:       logger,
		profileStore: store,
		runs:         make(map[string]*tunnelRun),
		snapshots:    make(map[string]*TunnelSnapshot),
		maxLogs:      200,
	}
}

func (c *TunnelController) Start(profileID string, cfg TunnelConfig) error {
	return c.start(profileID, cfg, nil)
}

func (c *TunnelController) StartWithListener(profileID string, cfg TunnelConfig, listener net.Listener) error {
	return c.start(profileID, cfg, listener)
}

func (c *TunnelController) start(profileID string, cfg TunnelConfig, listener net.Listener) error {
	if err := cfg.ValidateRuntime(); err != nil {
		return err
	}
	if strings.TrimSpace(profileID) == "" {
		return fmt.Errorf("profile ID is required")
	}

	c.Stop(profileID)

	ctx, cancel := context.WithCancel(context.Background())
	run := &tunnelRun{
		profileID: profileID,
		cfg:       cfg,
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
		listener:  listener,
		tracker:   newTunnelTracker(),
		sessions:  &sessionStore{},
		traffic:   &trafficMetrics{},
	}

	c.mu.Lock()
	c.runs[profileID] = run
	snapshot := c.snapshotLocked(profileID)
	snapshot.State = TunnelStateConnecting
	snapshot.Endpoint = cfg.LocalEndpoint()
	snapshot.Health = TunnelHealth{}
	snapshot.LastError = ""
	snapshot.Traffic = TrafficStats{}
	c.mu.Unlock()

	go c.executeRun(run)
	return nil
}

func (c *TunnelController) Stop(profileID string) {
	run := c.detachRun(profileID)
	if run == nil {
		c.mu.Lock()
		snapshot := c.snapshotLocked(profileID)
		snapshot.State = TunnelStateDisconnected
		snapshot.Health = TunnelHealth{}
		snapshot.Traffic.UploadBytesPerSec = 0
		snapshot.Traffic.DownloadBytesPerSec = 0
		snapshot.Traffic.ActiveConnections = 0
		c.mu.Unlock()
		return
	}

	run.cancel()
	if run.listener != nil {
		_ = run.listener.Close()
	}
	run.tracker.CloseAll()
	<-run.done
}

func (c *TunnelController) StopAll() {
	c.mu.RLock()
	ids := make([]string, 0, len(c.runs))
	for profileID := range c.runs {
		ids = append(ids, profileID)
	}
	c.mu.RUnlock()

	for _, profileID := range ids {
		c.Stop(profileID)
	}
}

func (c *TunnelController) Snapshot(profileID string) TunnelSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	snapshot, ok := c.snapshots[profileID]
	if !ok {
		return TunnelSnapshot{ProfileID: profileID, State: TunnelStateDisconnected}
	}
	cloned := cloneSnapshot(snapshot)
	if run := c.runs[profileID]; run != nil {
		cloned.Traffic = run.traffic.Snapshot()
	}
	return cloned
}

func (c *TunnelController) Snapshots() map[string]TunnelSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make(map[string]TunnelSnapshot, len(c.snapshots))
	for profileID, snapshot := range c.snapshots {
		cloned := cloneSnapshot(snapshot)
		if run := c.runs[profileID]; run != nil {
			cloned.Traffic = run.traffic.Snapshot()
		}
		out[profileID] = cloned
	}
	return out
}

func cloneSnapshot(snapshot *TunnelSnapshot) TunnelSnapshot {
	if snapshot == nil {
		return TunnelSnapshot{State: TunnelStateDisconnected}
	}
	logs := make([]LogEntry, len(snapshot.Logs))
	copy(logs, snapshot.Logs)
	return TunnelSnapshot{
		ProfileID: snapshot.ProfileID,
		State:     snapshot.State,
		Endpoint:  snapshot.Endpoint,
		Health:    snapshot.Health,
		LastError: snapshot.LastError,
		Traffic:   snapshot.Traffic,
		Logs:      logs,
	}
}

func (c *TunnelController) snapshotLocked(profileID string) *TunnelSnapshot {
	snapshot, ok := c.snapshots[profileID]
	if !ok {
		snapshot = &TunnelSnapshot{
			ProfileID: profileID,
			State:     TunnelStateDisconnected,
			Logs:      make([]LogEntry, 0, 32),
		}
		c.snapshots[profileID] = snapshot
	}
	return snapshot
}

func (c *TunnelController) detachRun(profileID string) *tunnelRun {
	c.mu.Lock()
	defer c.mu.Unlock()
	run := c.runs[profileID]
	if run != nil {
		delete(c.runs, profileID)
	}
	return run
}

func (c *TunnelController) executeRun(run *tunnelRun) {
	defer close(run.done)
	defer c.finalizeRun(run)

	go run.traffic.sampleLoop(run.ctx)

	if run.listener == nil {
		listener, err := net.Listen("tcp", run.cfg.LocalEndpoint())
		if err != nil {
			c.setLastError(run.profileID, err.Error())
			c.logf(run.profileID, "failed to bind local listener: %v", err)
			c.setState(run.profileID, TunnelStateDisconnected)
			return
		}
		run.listener = listener
	}

	boundPort, err := listenerPort(run.listener)
	if err == nil {
		c.setEndpoint(run.profileID, net.JoinHostPort(run.cfg.LocalAddr, strconv.Itoa(boundPort)))
	}
	c.setListenerHealth(run.profileID, true)
	c.logf(run.profileID, "listening on %s and forwarding to %s through %s", run.listener.Addr(), run.cfg.RemoteAddr(), run.cfg.SSHAddr())

	acceptErr := make(chan error, 1)
	go func() {
		acceptErr <- acceptLoop(run.ctx, run.listener, run.sessions, run.tracker, run.cfg, c, run.profileID)
	}()

	backoff := time.Second
	markedEnabled := false

	for {
		if err := run.ctx.Err(); err != nil {
			return
		}

		c.setState(run.profileID, TunnelStateConnecting)
		c.logf(run.profileID, "connecting to SSH jump host %s", run.cfg.SSHAddr())
		client, err := dialSSH(run.ctx, run.cfg)
		if err != nil {
			if errors.Is(err, errAuthFailure) {
				c.setLastError(run.profileID, err.Error())
				c.logf(run.profileID, "authentication failure: %v", err)
				c.setSSHHealth(run.profileID, false)
				c.setState(run.profileID, TunnelStateDisconnected)
				return
			}
			c.setLastError(run.profileID, err.Error())
			c.setSSHHealth(run.profileID, false)
			c.setState(run.profileID, TunnelStateRetrying)
			c.logf(run.profileID, "retryable SSH failure: %v; retrying in %s", err, backoff)
			if err := waitForRetry(run.ctx, backoff); err != nil {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}

		run.sessions.Set(client)
		c.setSSHHealth(run.profileID, true)
		c.setLastError(run.profileID, "")
		c.setState(run.profileID, TunnelStateConnected)
		c.logf(run.profileID, "SSH tunnel connected; forwarding local traffic to %s", run.cfg.RemoteAddr())
		if !markedEnabled && c.profileStore != nil && run.profileID != "" {
			now := time.Now()
			_ = c.profileStore.TouchLastEnabled(run.profileID, now)
			markedEnabled = true
		}

		keepAliveErr := monitorKeepAlive(run.ctx, client)
		select {
		case err := <-acceptErr:
			run.sessions.Clear(client)
			_ = client.Close()
			if err != nil {
				c.setLastError(run.profileID, err.Error())
				c.logf(run.profileID, "accept loop stopped: %v", err)
			}
			return
		case <-run.ctx.Done():
			run.sessions.Clear(client)
			_ = client.Close()
			return
		case err := <-keepAliveErr:
			run.sessions.Clear(client)
			_ = client.Close()
			run.tracker.CloseAll()
			run.traffic.ResetRates()
			c.setSSHHealth(run.profileID, false)
			c.setLastError(run.profileID, err.Error())
			c.setState(run.profileID, TunnelStateRetrying)
			c.logf(run.profileID, "SSH session lost: %v; reconnecting in %s", err, backoff)
			if err := waitForRetry(run.ctx, backoff); err != nil {
				return
			}
			backoff = nextBackoff(backoff)
		}
	}
}

func (c *TunnelController) finalizeRun(run *tunnelRun) {
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.runs[run.profileID]
	if current == run {
		delete(c.runs, run.profileID)
	}
	run.traffic.ResetLive()
	snapshot := c.snapshotLocked(run.profileID)
	snapshot.State = TunnelStateDisconnected
	snapshot.Health = TunnelHealth{}
	snapshot.Traffic = run.traffic.Snapshot()
}

func (c *TunnelController) logf(profileID string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if c.logger != nil {
		c.logger.Printf("[%s] %s", profileID, msg)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	snapshot := c.snapshotLocked(profileID)
	snapshot.Logs = append(snapshot.Logs, LogEntry{At: time.Now(), Message: msg})
	if len(snapshot.Logs) > c.maxLogs {
		snapshot.Logs = snapshot.Logs[len(snapshot.Logs)-c.maxLogs:]
	}
}

func (c *TunnelController) setState(profileID string, state TunnelState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshotLocked(profileID).State = state
}

func (c *TunnelController) setEndpoint(profileID, endpoint string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshotLocked(profileID).Endpoint = endpoint
}

func (c *TunnelController) setListenerHealth(profileID string, ready bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshotLocked(profileID).Health.ListenerReady = ready
}

func (c *TunnelController) setSSHHealth(profileID string, ready bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshotLocked(profileID).Health.SSHConnected = ready
}

func (c *TunnelController) setLastError(profileID, err string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshotLocked(profileID).LastError = err
}

func listenerPort(listener net.Listener) (int, error) {
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("listener address is not TCP")
	}
	return addr.Port, nil
}

func dialSSH(ctx context.Context, cfg TunnelConfig) (*ssh.Client, error) {
	authMethods, err := sshAuthMethods(cfg)
	if err != nil {
		return nil, err
	}

	hostKeyCallback, err := hostKeyCallback(cfg)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.SSHUser,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         cfg.ConnectTimeout,
	}

	dialCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", cfg.SSHAddr())
	if err != nil {
		return nil, classifyDialError(err)
	}

	if err := conn.SetDeadline(time.Now().Add(cfg.ConnectTimeout)); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("set handshake deadline: %w", err)
	}

	clientConn, chans, reqs, err := ssh.NewClientConn(conn, cfg.SSHAddr(), sshConfig)
	if err != nil {
		_ = conn.Close()
		return nil, classifyDialError(err)
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		_ = clientConn.Close()
		return nil, fmt.Errorf("clear handshake deadline: %w", err)
	}

	return ssh.NewClient(clientConn, chans, reqs), nil
}

func sshAuthMethods(cfg TunnelConfig) ([]ssh.AuthMethod, error) {
	methods := make([]ssh.AuthMethod, 0, 2)
	if cfg.SSHPassword != "" {
		return append(methods, ssh.Password(cfg.SSHPassword)), nil
	}
	if cfg.SSHKeyPath != "" {
		keyPath, err := expandPath(cfg.SSHKeyPath)
		if err != nil {
			return nil, err
		}
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("read SSH key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("parse SSH key: %w", err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	return methods, nil
}

func hostKeyCallback(cfg TunnelConfig) (ssh.HostKeyCallback, error) {
	if cfg.InsecureIgnoreHost {
		return ssh.InsecureIgnoreHostKey(), nil
	}
	if cfg.KnownHostsPath == "" {
		return nil, errors.New("known_hosts path is empty; set known_hosts or use insecure ignore host key")
	}
	knownHostsPath, err := expandPath(cfg.KnownHostsPath)
	if err != nil {
		return nil, err
	}
	return knownhosts.New(knownHostsPath)
}

var errAuthFailure = errors.New("authentication failure")

func classifyDialError(err error) error {
	var authErr *ssh.ServerAuthError
	if errors.As(err, &authErr) {
		return fmt.Errorf("%w: %v", errAuthFailure, err)
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return fmt.Errorf("network timeout: %w", err)
		}
		return fmt.Errorf("network error: %w", err)
	}

	errText := strings.ToLower(err.Error())
	if strings.Contains(errText, "unable to authenticate") || strings.Contains(errText, "permission denied") {
		return fmt.Errorf("%w: %v", errAuthFailure, err)
	}

	return fmt.Errorf("retryable SSH error: %w", err)
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func nextBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > maxBackoff {
		return maxBackoff
	}
	return next
}

func monitorKeepAlive(ctx context.Context, client *ssh.Client) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		ticker := time.NewTicker(keepAliveInterval)
		defer ticker.Stop()
		defer close(errCh)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := keepAliveOnce(client); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()

	return errCh
}

func keepAliveOnce(client *ssh.Client) error {
	result := make(chan error, 1)

	go func() {
		_, _, err := client.Conn.SendRequest("keepalive@openssh.com", true, nil)
		result <- err
	}()

	select {
	case err := <-result:
		if err != nil {
			return fmt.Errorf("keepalive failed: %w", err)
		}
		return nil
	case <-time.After(keepAliveTimeout):
		return fmt.Errorf("keepalive timed out after %s", keepAliveTimeout)
	}
}

func acceptLoop(
	ctx context.Context,
	listener net.Listener,
	sessions *sessionStore,
	tracker *tunnelTracker,
	cfg TunnelConfig,
	controller *TunnelController,
	profileID string,
) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Temporary() {
				controller.logf(profileID, "temporary accept error: %v", err)
				continue
			}
			return fmt.Errorf("accept local connection: %w", err)
		}

		wg.Add(1)
		go func(local net.Conn) {
			defer wg.Done()
			handleLocalConn(local, sessions, tracker, cfg, controller, profileID)
		}(conn)
	}
}

func handleLocalConn(
	local net.Conn,
	sessions *sessionStore,
	tracker *tunnelTracker,
	cfg TunnelConfig,
	controller *TunnelController,
	profileID string,
) {
	client := sessions.Get()
	if client == nil {
		controller.logf(profileID, "rejecting local client %s because SSH session is unavailable", local.RemoteAddr())
		_ = local.Close()
		return
	}

	remote, err := client.Dial("tcp", cfg.RemoteAddr())
	if err != nil {
		controller.logf(profileID, "failed to open remote connection for local client %s: %v", local.RemoteAddr(), err)
		_ = local.Close()
		return
	}

	tunnel := &trackedTunnel{local: local, remote: remote}
	tracker.Add(tunnel)
	runTraffic := controller.lookupTraffic(profileID)
	if runTraffic != nil {
		runTraffic.IncActive()
	}
	defer func() {
		tunnel.Close()
		tracker.Remove(tunnel)
		if runTraffic != nil {
			runTraffic.DecActive()
		}
	}()

	copyDone := make(chan struct{}, 2)
	go proxy(local, remote, copyDone, func(n int) {
		if runTraffic != nil {
			runTraffic.AddUpload(n)
		}
	})
	go proxy(remote, local, copyDone, func(n int) {
		if runTraffic != nil {
			runTraffic.AddDownload(n)
		}
	})

	<-copyDone
	tunnel.Close()
	<-copyDone
}

func (c *TunnelController) lookupTraffic(profileID string) *trafficMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	run := c.runs[profileID]
	if run == nil {
		return nil
	}
	return run.traffic
}

type countingReader struct {
	reader  io.Reader
	observe func(int)
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 && r.observe != nil {
		r.observe(n)
	}
	return n, err
}

func proxy(src net.Conn, dst net.Conn, done chan<- struct{}, observe func(int)) {
	defer func() { done <- struct{}{} }()
	_, _ = io.Copy(dst, &countingReader{reader: src, observe: observe})
	closeWrite(dst)
	closeRead(src)
}

func closeWrite(conn net.Conn) {
	type closeWriter interface {
		CloseWrite() error
	}
	if c, ok := conn.(closeWriter); ok {
		_ = c.CloseWrite()
		return
	}
	_ = conn.Close()
}

func closeRead(conn net.Conn) {
	type closeReader interface {
		CloseRead() error
	}
	if c, ok := conn.(closeReader); ok {
		_ = c.CloseRead()
	}
}
