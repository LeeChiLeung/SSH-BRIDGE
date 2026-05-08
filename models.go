package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	keepAliveInterval = 5 * time.Second
	keepAliveTimeout  = 15 * time.Second
	maxBackoff        = 60 * time.Second
	envVarName        = "TUNNEL_LOCAL_PORT"
	defaultSSHHost    = "ssh.example.com"
	defaultSSHPort    = 22
	defaultSSHUser    = "user"
	defaultRemoteHost = "127.0.0.1"
	defaultRemotePort = 5432
	defaultLocalAddr  = "127.0.0.1"
	defaultLocalPort  = 15432
	defaultWebAddr    = "127.0.0.1:8848"
	defaultTitle      = "示例隧道"
)

type TunnelConfig struct {
	Title              string        `json:"title"`
	SSHHost            string        `json:"sshHost"`
	SSHPort            int           `json:"sshPort"`
	SSHUser            string        `json:"sshUser"`
	SSHPassword        string        `json:"-"`
	SSHKeyPath         string        `json:"sshKeyPath"`
	RemoteHost         string        `json:"remoteHost"`
	RemotePort         int           `json:"remotePort"`
	LocalAddr          string        `json:"localAddr"`
	LocalPort          int           `json:"localPort"`
	ConnectTimeout     time.Duration `json:"-"`
	KnownHostsPath     string        `json:"knownHostsPath"`
	InsecureIgnoreHost bool          `json:"insecureIgnoreHost"`
}

func (c TunnelConfig) SSHAddr() string {
	return net.JoinHostPort(c.SSHHost, strconv.Itoa(c.SSHPort))
}

func (c TunnelConfig) RemoteAddr() string {
	return net.JoinHostPort(c.RemoteHost, strconv.Itoa(c.RemotePort))
}

func (c TunnelConfig) LocalEndpoint() string {
	return net.JoinHostPort(c.LocalAddr, strconv.Itoa(c.LocalPort))
}

func (c TunnelConfig) ValidateRuntime() error {
	if c.SSHHost == "" {
		return fmt.Errorf("SSH host is required")
	}
	if c.SSHUser == "" {
		return fmt.Errorf("SSH user is required")
	}
	if c.SSHPort <= 0 || c.SSHPort > 65535 {
		return fmt.Errorf("SSH port must be between 1 and 65535")
	}
	if c.LocalPort < 0 || c.LocalPort > 65535 {
		return fmt.Errorf("local port must be between 0 and 65535")
	}
	if c.RemoteHost == "" {
		return fmt.Errorf("remote host is required")
	}
	if c.RemotePort <= 0 || c.RemotePort > 65535 {
		return fmt.Errorf("remote port must be between 1 and 65535")
	}
	if c.LocalAddr == "" {
		return fmt.Errorf("local address is required")
	}
	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("connect timeout must be positive")
	}
	if c.SSHPassword == "" && c.SSHKeyPath == "" {
		return fmt.Errorf("either SSH password or private key path is required")
	}
	return nil
}

func (c TunnelConfig) DisplayTitle() string {
	if strings.TrimSpace(c.Title) == "" {
		return defaultTitle
	}
	return c.Title
}

func defaultTunnelConfig() TunnelConfig {
	return TunnelConfig{
		Title:              defaultTitle,
		SSHHost:            defaultSSHHost,
		SSHPort:            defaultSSHPort,
		SSHUser:            defaultSSHUser,
		RemoteHost:         defaultRemoteHost,
		RemotePort:         defaultRemotePort,
		LocalAddr:          defaultLocalAddr,
		LocalPort:          defaultLocalPort,
		ConnectTimeout:     10 * time.Second,
		KnownHostsPath:     defaultKnownHostsPath(),
		InsecureIgnoreHost: false,
	}
}

func defaultKnownHostsPath() string {
	home, err := osUserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "known_hosts")
}

func parseHostPort(raw string) (string, int, error) {
	host, portText, err := net.SplitHostPort(raw)
	if err != nil {
		return "", 0, err
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func osUserHomeDir() (string, error) {
	return os.UserHomeDir()
}
