//go:build linux || windows || darwin

package main

import (
	"github.com/energye/systray"
)

type systrayManager struct{}

type systrayMenuItem struct {
	item *systray.MenuItem
}

func newDefaultTrayManager() (trayManager, error) {
	return systrayManager{}, nil
}

func (systrayManager) Run(onReady func(), onExit func()) {
	systray.Run(onReady, onExit)
}

func (systrayManager) Quit() {
	systray.Quit()
}

func (systrayManager) SetTitle(title string) {
	systray.SetTitle(title)
}

func (systrayManager) SetTooltip(tooltip string) {
	systray.SetTooltip(tooltip)
}

func (systrayManager) SetIcon(icon []byte) {
	systray.SetIcon(icon)
}

func (systrayManager) AddMenuItem(title string, tooltip string) trayMenuItem {
	return systrayMenuItem{item: systray.AddMenuItem(title, tooltip)}
}

func (i systrayMenuItem) OnClick(fn func()) {
	i.item.Click(fn)
}
