package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

type cliOptions struct {
	TunnelConfig
	EnvMode    bool
	ChildMode  bool
	ListenerFD int
	WebMode    bool
	WebAddr    string
}

func main() {
	opts, err := parseCLI()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)

	if opts.WebMode {
		if err := runWebMode(opts, logger); err != nil {
			logger.Printf("web console exited with error: %v", err)
			os.Exit(1)
		}
		return
	}

	if err := runTunnelMode(opts, logger); err != nil {
		logger.Printf("tunnel exited with error: %v", err)
		os.Exit(1)
	}
}

func parseCLI() (cliOptions, error) {
	cfg := defaultTunnelConfig()
	opts := cliOptions{TunnelConfig: cfg, WebAddr: defaultWebAddr}

	var sshAddrOverride string
	var remoteAddrOverride string

	flag.StringVar(&opts.Title, "title", cfg.Title, "Display title for the tunnel profile")
	flag.StringVar(&opts.SSHHost, "ssh-host", cfg.SSHHost, "SSH jump server host")
	flag.IntVar(&opts.SSHPort, "ssh-port", cfg.SSHPort, "SSH jump server port")
	flag.StringVar(&sshAddrOverride, "ssh-addr", "", "Legacy host:port SSH jump server address override")
	flag.StringVar(&opts.SSHUser, "ssh-user", cfg.SSHUser, "SSH username")
	flag.StringVar(&opts.SSHPassword, "ssh-password", "", "SSH password")
	flag.StringVar(&opts.SSHKeyPath, "ssh-key", "", "Path to a PEM-encoded private key")
	flag.StringVar(&opts.RemoteHost, "remote-host", cfg.RemoteHost, "Remote TCP target host")
	flag.IntVar(&opts.RemotePort, "remote-port", cfg.RemotePort, "Remote TCP target port")
	flag.StringVar(&remoteAddrOverride, "remote-addr", "", "Legacy host:port remote TCP target override")
	flag.StringVar(&opts.LocalAddr, "local-addr", cfg.LocalAddr, "Local bind address")
	flag.IntVar(&opts.LocalPort, "local-port", cfg.LocalPort, "Local port to bind; use 0 for an ephemeral port")
	flag.DurationVar(&opts.ConnectTimeout, "connect-timeout", cfg.ConnectTimeout, "SSH TCP connect and handshake timeout")
	flag.StringVar(&opts.KnownHostsPath, "known-hosts", cfg.KnownHostsPath, "Path to known_hosts file")
	flag.BoolVar(&opts.InsecureIgnoreHost, "insecure-ignore-host-key", false, "Disable SSH host key verification")
	flag.BoolVar(&opts.EnvMode, "env", false, "Print export command and keep the tunnel running in the background")
	flag.BoolVar(&opts.ChildMode, "child", false, "Internal flag used by --env mode")
	flag.IntVar(&opts.ListenerFD, "listener-fd", 0, "Internal file descriptor used by --env mode")
	flag.BoolVar(&opts.WebMode, "web", false, "Run the local Web control console instead of starting the tunnel directly")
	flag.StringVar(&opts.WebAddr, "web-addr", defaultWebAddr, "Local Web console listen address")
	flag.Parse()

	if sshAddrOverride != "" {
		host, port, err := parseHostPort(sshAddrOverride)
		if err != nil {
			return opts, fmt.Errorf("parse --ssh-addr: %w", err)
		}
		opts.SSHHost = host
		opts.SSHPort = port
	}
	if remoteAddrOverride != "" {
		host, port, err := parseHostPort(remoteAddrOverride)
		if err != nil {
			return opts, fmt.Errorf("parse --remote-addr: %w", err)
		}
		opts.RemoteHost = host
		opts.RemotePort = port
	}

	if opts.WebMode {
		if opts.WebAddr == "" {
			return opts, fmt.Errorf("web listen address is required")
		}
		return opts, nil
	}
	if opts.ChildMode && opts.ListenerFD < 3 {
		return opts, fmt.Errorf("child mode requires a valid --listener-fd")
	}
	if err := opts.TunnelConfig.ValidateRuntime(); err != nil {
		return opts, err
	}
	return opts, nil
}

func runTunnelMode(opts cliOptions, logger *log.Logger) error {
	listener, err := prepareListener(opts)
	if err != nil {
		return err
	}

	if opts.EnvMode && !opts.ChildMode {
		if err := spawnChildAndPrintEnv(opts, listener); err != nil {
			_ = listener.Close()
			return err
		}
		return nil
	}

	controller := NewTunnelController(logger, nil)
	if err := controller.StartWithListener("cli", opts.TunnelConfig, listener); err != nil {
		_ = listener.Close()
		return err
	}
	defer controller.Stop()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	return nil
}

func runWebMode(opts cliOptions, logger *log.Logger) error {
	store, err := NewProfileStore(defaultProfileStorePath())
	if err != nil {
		return err
	}
	if _, err := store.EnsureDefault(opts.TunnelConfig); err != nil {
		return err
	}

	controller := NewTunnelController(logger, store)
	server, err := NewWebServer(opts.WebAddr, controller, store, logger)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return server.Run(ctx)
}

func prepareListener(opts cliOptions) (net.Listener, error) {
	if opts.ChildMode {
		file := os.NewFile(uintptr(opts.ListenerFD), "tunnel-listener")
		defer file.Close()
		return net.FileListener(file)
	}
	return net.Listen("tcp", opts.LocalEndpoint())
}

func spawnChildAndPrintEnv(opts cliOptions, listener net.Listener) error {
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		return errors.New("listener is not a TCP listener")
	}

	file, err := tcpListener.File()
	if err != nil {
		return fmt.Errorf("duplicate listener file descriptor: %w", err)
	}
	defer file.Close()

	boundPort, err := listenerPort(listener)
	if err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	childArgs := buildChildArgs(opts, boundPort)
	cmd := exec.Command(exe, childArgs...)
	cmd.ExtraFiles = []*os.File{file}
	cmd.Stdin = nil
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start child process: %w", err)
	}

	_ = listener.Close()
	fmt.Printf("export %s=%d\n", envVarName, boundPort)
	return nil
}

func buildChildArgs(opts cliOptions, boundPort int) []string {
	args := []string{
		"--title", opts.DisplayTitle(),
		"--ssh-host", opts.SSHHost,
		"--ssh-port", strconv.Itoa(opts.SSHPort),
		"--ssh-user", opts.SSHUser,
		"--remote-host", opts.RemoteHost,
		"--remote-port", strconv.Itoa(opts.RemotePort),
		"--local-addr", opts.LocalAddr,
		"--local-port", strconv.Itoa(boundPort),
		"--connect-timeout", opts.ConnectTimeout.String(),
		"--child",
		"--listener-fd", "3",
	}
	if opts.SSHPassword != "" {
		args = append(args, "--ssh-password", opts.SSHPassword)
	}
	if opts.SSHKeyPath != "" {
		args = append(args, "--ssh-key", opts.SSHKeyPath)
	}
	if opts.KnownHostsPath != "" {
		args = append(args, "--known-hosts", opts.KnownHostsPath)
	}
	if opts.InsecureIgnoreHost {
		args = append(args, "--insecure-ignore-host-key")
	}
	return args
}
