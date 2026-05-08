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

type TunnelSnapshot struct {
	State           TunnelState  `json:"state"`
	ActiveProfileID string       `json:"activeProfileId"`
	Endpoint        string       `json:"endpoint"`
	Health          TunnelHealth `json:"health"`
	LastError       string       `json:"lastError"`
	Logs            []LogEntry   `json:"logs"`
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
}

type TunnelController struct {
	mu              sync.RWMutex
	logger          *log.Logger
	profileStore    *ProfileStore
	run             *tunnelRun
	state           TunnelState
	activeProfileID string
	endpoint        string
	health          TunnelHealth
	lastError       string
	logs            []LogEntry
	maxLogs         int
}

func NewTunnelController(logger *log.Logger, store *ProfileStore) *TunnelController {
	return &TunnelController{
		logger:       logger,
		state:        TunnelStateDisconnected,
		maxLogs:      200,
		logs:         make([]LogEntry, 0, 64),
		profileStore: store,
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
	c.Stop()

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
	}

	c.mu.Lock()
	c.run = run
	c.activeProfileID = profileID
	c.state = TunnelStateConnecting
	c.endpoint = cfg.LocalEndpoint()
	c.health = TunnelHealth{}
	c.lastError = ""
	c.mu.Unlock()

	go c.executeRun(run)
	return nil
}

func (c *TunnelController) Stop() {
	c.mu.Lock()
	run := c.run
	if run == nil {
		c.state = TunnelStateDisconnected
		c.health = TunnelHealth{}
		c.endpoint = ""
		c.lastError = ""
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	run.cancel()
	if run.listener != nil {
		_ = run.listener.Close()
	}
	run.tracker.CloseAll()
	<-run.done
}

func (c *TunnelController) Snapshot() TunnelSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	logs := make([]LogEntry, len(c.logs))
	copy(logs, c.logs)
	return TunnelSnapshot{
		State:           c.state,
		ActiveProfileID: c.activeProfileID,
		Endpoint:        c.endpoint,
		Health:          c.health,
		LastError:       c.lastError,
		Logs:            logs,
	}
}

func (c *TunnelController) executeRun(run *tunnelRun) {
	defer close(run.done)
	defer c.finalizeRun(run)

	if run.listener == nil {
		listener, err := net.Listen("tcp", run.cfg.LocalEndpoint())
		if err != nil {
			c.setLastError(err.Error())
			c.logf("failed to bind local listener: %v", err)
			c.setState(TunnelStateDisconnected)
			return
		}
		run.listener = listener
	}

	boundPort, err := listenerPort(run.listener)
	if err == nil {
		c.setEndpoint(net.JoinHostPort(run.cfg.LocalAddr, strconv.Itoa(boundPort)))
	}
	c.setListenerHealth(true)
	c.logf("listening on %s and forwarding to %s through %s", run.listener.Addr(), run.cfg.RemoteAddr(), run.cfg.SSHAddr())

	acceptErr := make(chan error, 1)
	go func() {
		acceptErr <- acceptLoop(run.ctx, run.listener, run.sessions, run.tracker, run.cfg, c)
	}()

	backoff := time.Second
	markedEnabled := false

	for {
		if err := run.ctx.Err(); err != nil {
			return
		}

		c.setState(TunnelStateConnecting)
		c.logf("connecting to SSH jump host %s", run.cfg.SSHAddr())
		client, err := dialSSH(run.ctx, run.cfg)
		if err != nil {
			if errors.Is(err, errAuthFailure) {
				c.setLastError(err.Error())
				c.logf("authentication failure: %v", err)
				c.setSSHHealth(false)
				c.setState(TunnelStateDisconnected)
				return
			}
			c.setLastError(err.Error())
			c.setSSHHealth(false)
			c.setState(TunnelStateRetrying)
			c.logf("retryable SSH failure: %v; retrying in %s", err, backoff)
			if err := waitForRetry(run.ctx, backoff); err != nil {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}

		run.sessions.Set(client)
		c.setSSHHealth(true)
		c.setLastError("")
		c.setState(TunnelStateConnected)
		c.logf("SSH tunnel connected; forwarding local traffic to %s", run.cfg.RemoteAddr())
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
				c.setLastError(err.Error())
				c.logf("accept loop stopped: %v", err)
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
			c.setSSHHealth(false)
			c.setLastError(err.Error())
			c.setState(TunnelStateRetrying)
			c.logf("SSH session lost: %v; reconnecting in %s", err, backoff)
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
	if c.run == run {
		c.run = nil
		c.state = TunnelStateDisconnected
		c.health = TunnelHealth{}
	}
}

func (c *TunnelController) logf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if c.logger != nil {
		c.logger.Print(msg)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logs = append(c.logs, LogEntry{At: time.Now(), Message: msg})
	if len(c.logs) > c.maxLogs {
		c.logs = c.logs[len(c.logs)-c.maxLogs:]
	}
}

func (c *TunnelController) setState(state TunnelState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = state
}

func (c *TunnelController) setEndpoint(endpoint string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.endpoint = endpoint
}

func (c *TunnelController) setListenerHealth(ready bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.health.ListenerReady = ready
}

func (c *TunnelController) setSSHHealth(ready bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.health.SSHConnected = ready
}

func (c *TunnelController) setLastError(err string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastError = err
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
		methods = append(methods, ssh.Password(cfg.SSHPassword))
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
				controller.logf("temporary accept error: %v", err)
				continue
			}
			return fmt.Errorf("accept local connection: %w", err)
		}

		wg.Add(1)
		go func(local net.Conn) {
			defer wg.Done()
			handleLocalConn(local, sessions, tracker, cfg, controller)
		}(conn)
	}
}

func handleLocalConn(
	local net.Conn,
	sessions *sessionStore,
	tracker *tunnelTracker,
	cfg TunnelConfig,
	controller *TunnelController,
) {
	client := sessions.Get()
	if client == nil {
		controller.logf("rejecting local client %s because SSH session is unavailable", local.RemoteAddr())
		_ = local.Close()
		return
	}

	remote, err := client.Dial("tcp", cfg.RemoteAddr())
	if err != nil {
		controller.logf("failed to open remote connection for local client %s: %v", local.RemoteAddr(), err)
		_ = local.Close()
		return
	}

	tunnel := &trackedTunnel{local: local, remote: remote}
	tracker.Add(tunnel)
	defer func() {
		tunnel.Close()
		tracker.Remove(tunnel)
	}()

	copyDone := make(chan struct{}, 2)
	go proxy(local, remote, copyDone)
	go proxy(remote, local, copyDone)

	<-copyDone
	tunnel.Close()
	<-copyDone
}

func proxy(src net.Conn, dst net.Conn, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()
	_, _ = io.Copy(dst, src)
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
