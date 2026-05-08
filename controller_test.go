package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestTunnelControllerSupportsConcurrentProfiles(t *testing.T) {
	t.Parallel()

	sshAddr, closeSSH := startTestSSHServer(t, "secret")
	defer closeSSH()

	echoOne, closeEchoOne := startEchoServer(t, "alpha")
	defer closeEchoOne()
	echoTwo, closeEchoTwo := startEchoServer(t, "bravo")
	defer closeEchoTwo()

	controller := NewTunnelController(log.New(io.Discard, "", 0), newTestProfileStore(t))

	cfgOne := testTunnelConfig(sshAddr, echoOne, "secret")
	cfgTwo := testTunnelConfig(sshAddr, echoTwo, "secret")

	if err := controller.Start("one", cfgOne); err != nil {
		t.Fatalf("start one: %v", err)
	}
	if err := controller.Start("two", cfgTwo); err != nil {
		t.Fatalf("start two: %v", err)
	}
	defer controller.StopAll()

	waitForState(t, controller, "one", TunnelStateConnected)
	waitForState(t, controller, "two", TunnelStateConnected)

	assertForwarding(t, controller.Snapshot("one").Endpoint, "alpha")
	assertForwarding(t, controller.Snapshot("two").Endpoint, "bravo")

	controller.Stop("one")
	waitForState(t, controller, "one", TunnelStateDisconnected)
	waitForState(t, controller, "two", TunnelStateConnected)
	assertForwarding(t, controller.Snapshot("two").Endpoint, "bravo")
}

func TestTunnelControllerIsolatesRetryingAndPortConflicts(t *testing.T) {
	t.Parallel()

	sshAddr, closeSSH := startTestSSHServer(t, "secret")
	defer closeSSH()

	echoAddr, closeEcho := startEchoServer(t, "steady")
	defer closeEcho()

	controller := NewTunnelController(log.New(io.Discard, "", 0), newTestProfileStore(t))
	defer controller.StopAll()

	healthyCfg := testTunnelConfig(sshAddr, echoAddr, "secret")
	if err := controller.Start("healthy", healthyCfg); err != nil {
		t.Fatalf("start healthy: %v", err)
	}
	waitForState(t, controller, "healthy", TunnelStateConnected)

	retryingCfg := healthyCfg
	retryingCfg.SSHPort = unusedPort(t)
	retryingCfg.LocalPort = 0
	if err := controller.Start("retrying", retryingCfg); err != nil {
		t.Fatalf("start retrying: %v", err)
	}
	waitForState(t, controller, "retrying", TunnelStateRetrying)

	conflictCfg := healthyCfg
	conflictCfg.LocalPort = localPortFromEndpoint(t, controller.Snapshot("healthy").Endpoint)
	if err := controller.Start("conflict", conflictCfg); err != nil {
		t.Fatalf("start conflict: %v", err)
	}
	waitForState(t, controller, "conflict", TunnelStateDisconnected)

	healthy := controller.Snapshot("healthy")
	retrying := controller.Snapshot("retrying")
	conflict := controller.Snapshot("conflict")

	if healthy.State != TunnelStateConnected {
		t.Fatalf("healthy state = %s, want connected", healthy.State)
	}
	if retrying.State != TunnelStateRetrying {
		t.Fatalf("retrying state = %s, want retrying", retrying.State)
	}
	if !strings.Contains(conflict.LastError, "address already in use") && !strings.Contains(conflict.LastError, "bind") {
		t.Fatalf("conflict last error = %q, want bind conflict", conflict.LastError)
	}

	assertForwarding(t, healthy.Endpoint, "steady")
}

func TestWebServerBuildProfileViewsUsesProfileSnapshots(t *testing.T) {
	t.Parallel()

	store := newTestProfileStore(t)
	profileOne, err := store.Save(profileFromConfig("one", defaultTunnelConfig()))
	if err != nil {
		t.Fatalf("save one: %v", err)
	}
	cfgTwo := defaultTunnelConfig()
	cfgTwo.LocalPort = 4180
	profileTwo, err := store.Save(profileFromConfig("two", cfgTwo))
	if err != nil {
		t.Fatalf("save two: %v", err)
	}

	controller := NewTunnelController(log.New(io.Discard, "", 0), store)
	controller.mu.Lock()
	controller.snapshots["one"] = &TunnelSnapshot{
		ProfileID: "one",
		State:     TunnelStateConnected,
		Endpoint:  "127.0.0.1:3180",
		Health:    TunnelHealth{ListenerReady: true, SSHConnected: true},
	}
	controller.snapshots["two"] = &TunnelSnapshot{
		ProfileID: "two",
		State:     TunnelStateRetrying,
		Endpoint:  "127.0.0.1:4180",
		Health:    TunnelHealth{ListenerReady: true, SSHConnected: false},
		LastError: "network timeout",
	}
	controller.mu.Unlock()

	server := &WebServer{controller: controller, store: store}
	views := server.buildProfileViews()
	if len(views) != 2 {
		t.Fatalf("views len = %d, want 2", len(views))
	}

	viewMap := map[string]profileView{
		views[0].Profile.ID: views[0],
		views[1].Profile.ID: views[1],
	}
	if got := viewMap[profileOne.ID].State; got != TunnelStateConnected {
		t.Fatalf("profile one state = %s, want connected", got)
	}
	if got := viewMap[profileTwo.ID].State; got != TunnelStateRetrying {
		t.Fatalf("profile two state = %s, want retrying", got)
	}
	if got := viewMap[profileTwo.ID].LastError; got != "network timeout" {
		t.Fatalf("profile two last error = %q", got)
	}
}

func newTestProfileStore(t *testing.T) *ProfileStore {
	t.Helper()
	store, err := NewProfileStore(filepath.Join(t.TempDir(), "profiles.json"))
	if err != nil {
		t.Fatalf("new profile store: %v", err)
	}
	return store
}

func waitForState(t *testing.T, controller *TunnelController, profileID string, want TunnelState) {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if controller.Snapshot(profileID).State == want {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	snapshot := controller.Snapshot(profileID)
	t.Fatalf("profile %s state = %s, want %s, last error = %q", profileID, snapshot.State, want, snapshot.LastError)
}

func assertForwarding(t *testing.T, endpoint string, want string) {
	t.Helper()
	conn, err := net.DialTimeout("tcp", endpoint, 2*time.Second)
	if err != nil {
		t.Fatalf("dial local endpoint %s: %v", endpoint, err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatalf("write local endpoint %s: %v", endpoint, err)
	}

	buf := make([]byte, len(want))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read local endpoint %s: %v", endpoint, err)
	}
	if got := string(buf); got != want {
		t.Fatalf("forwarded response = %q, want %q", got, want)
	}
}

func testTunnelConfig(sshAddr, remoteAddr, password string) TunnelConfig {
	sshHost, sshPort, _ := parseHostPort(sshAddr)
	remoteHost, remotePort, _ := parseHostPort(remoteAddr)
	return TunnelConfig{
		Title:              "test",
		SSHHost:            sshHost,
		SSHPort:            sshPort,
		SSHUser:            "tester",
		SSHPassword:        password,
		RemoteHost:         remoteHost,
		RemotePort:         remotePort,
		LocalAddr:          "127.0.0.1",
		LocalPort:          0,
		ConnectTimeout:     2 * time.Second,
		InsecureIgnoreHost: true,
	}
}

func localPortFromEndpoint(t *testing.T, endpoint string) int {
	t.Helper()
	_, port, err := parseHostPort(endpoint)
	if err != nil {
		t.Fatalf("parse endpoint %s: %v", endpoint, err)
	}
	return port
}

func unusedPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func startEchoServer(t *testing.T, reply string) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen echo: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 4)
				_, _ = io.ReadFull(c, buf)
				_, _ = c.Write([]byte(reply))
			}(conn)
		}
	}()

	return listener.Addr().String(), func() {
		_ = listener.Close()
		<-done
	}
}

func startTestSSHServer(t *testing.T, password string) (string, func()) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate ssh host key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if conn.User() == "tester" && string(pass) == password {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected")
		},
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen ssh: %v", err)
	}

	var wg sync.WaitGroup
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			rawConn, err := listener.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func(conn net.Conn) {
				defer wg.Done()
				handleSSHConn(conn, config)
			}(rawConn)
		}
	}()

	return listener.Addr().String(), func() {
		_ = listener.Close()
		<-done
		wg.Wait()
	}
}

func handleSSHConn(rawConn net.Conn, config *ssh.ServerConfig) {
	serverConn, chans, reqs, err := ssh.NewServerConn(rawConn, config)
	if err != nil {
		_ = rawConn.Close()
		return
	}
	defer serverConn.Close()

	go ssh.DiscardRequests(reqs)
	for newChannel := range chans {
		if newChannel.ChannelType() != "direct-tcpip" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}

		var payload struct {
			DestAddr   string
			DestPort   uint32
			OriginAddr string
			OriginPort uint32
		}
		if err := ssh.Unmarshal(newChannel.ExtraData(), &payload); err != nil {
			_ = newChannel.Reject(ssh.ConnectionFailed, "invalid payload")
			continue
		}

		target := net.JoinHostPort(payload.DestAddr, fmt.Sprintf("%d", payload.DestPort))
		upstream, err := net.DialTimeout("tcp", target, 2*time.Second)
		if err != nil {
			_ = newChannel.Reject(ssh.ConnectionFailed, err.Error())
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			_ = upstream.Close()
			continue
		}
		go ssh.DiscardRequests(requests)

		go func() {
			defer upstream.Close()
			defer channel.Close()
			var copyWG sync.WaitGroup
			copyWG.Add(2)
			go func() {
				defer copyWG.Done()
				_, _ = io.Copy(channel, upstream)
			}()
			go func() {
				defer copyWG.Done()
				_, _ = io.Copy(upstream, channel)
			}()
			copyWG.Wait()
		}()
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
