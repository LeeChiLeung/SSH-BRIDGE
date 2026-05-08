//go:build !(linux || windows || darwin)

package main

import "fmt"

func newDefaultTrayManager() (trayManager, error) {
	return nil, fmt.Errorf("system tray mode is not supported on this platform")
}
