package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	_ "embed"
)

//go:embed icon/32x32.png
var embeddedCustomTrayIcon []byte

const (
	trayOpenNewConsole   = "打开控制台"
	trayQuitLabel        = "退出"
)

type trayMenuItem interface {
	OnClick(fn func())
}

type trayManager interface {
	Run(onReady func(), onExit func())
	Quit()
	SetTitle(title string)
	SetTooltip(tooltip string)
	SetIcon(icon []byte)
	AddMenuItem(title string, tooltip string) trayMenuItem
}

type browserOpener interface {
	Open(url string) error
}

type trayApp struct {
	manager   trayManager
	opener    browserOpener
	logger    *log.Logger
	newWebURL string
	onQuit    func()
}

func newDefaultTrayApp(logger *log.Logger, onOpenNewURL string, onQuit func()) (*trayApp, error) {
	manager, err := newDefaultTrayManager()
	if err != nil {
		return nil, err
	}
	return &trayApp{
		manager:   manager,
		opener:    commandBrowserOpener{},
		logger:    logger,
		newWebURL: onOpenNewURL,
		onQuit:    onQuit,
	}, nil
}

func (a *trayApp) Run(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		<-ctx.Done()
		a.manager.Quit()
	}()

	a.manager.Run(func() {
		a.manager.SetTitle("Timber SSH Go")
		a.manager.SetTooltip("Timber SSH Go")
		a.manager.SetIcon(resolveTrayIcon())

		newOpenItem := a.manager.AddMenuItem(trayOpenNewConsole, "打开新的本地 Web 控制台")
		quitItem := a.manager.AddMenuItem(trayQuitLabel, "退出程序")

		newOpenItem.OnClick(func() {
			a.openURL(a.newWebURL)
		})
		quitItem.OnClick(func() {
			if a.onQuit != nil {
				a.onQuit()
			}
			a.manager.Quit()
		})
	}, func() {
		close(done)
	})

	<-done
	return nil
}

func (a *trayApp) openURL(url string) {
	if url == "" {
		return
	}
	if err := a.opener.Open(url); err != nil && a.logger != nil {
		a.logger.Printf("open browser failed: %v", err)
	}
}

func resolveTrayIcon() []byte {
	for _, path := range trayIconCandidates() {
		iconData, err := loadTrayIcon(path)
		if err == nil && len(iconData) > 0 {
			return iconData
		}
	}
	if len(embeddedCustomTrayIcon) > 0 {
		return embeddedCustomTrayIcon
	}
	return embeddedTrayIcon()
}

func trayIconCandidates() []string {
	paths := []string{filepath.Join("icon", "32x32.png")}
	exePath, err := os.Executable()
	if err != nil {
		return paths
	}

	exeDir := filepath.Dir(exePath)
	paths = append(paths, filepath.Join(exeDir, "icon", "32x32.png"))
	paths = append(paths, filepath.Join(exeDir, "..", "share", "icons", "hicolor", "32x32", "apps", "timber-ssh-go.png"))
	return paths
}

func loadTrayIcon(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func embeddedTrayIcon() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	bg := color.NRGBA{R: 15, G: 118, B: 110, A: 255}
	fg := color.NRGBA{R: 248, G: 250, B: 252, A: 255}

	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetNRGBA(x, y, bg)
		}
	}
	for y := 4; y < 12; y++ {
		img.SetNRGBA(4, y, fg)
		img.SetNRGBA(11, y, fg)
	}
	for x := 4; x < 12; x++ {
		img.SetNRGBA(x, 4, fg)
		img.SetNRGBA(x, 11, fg)
	}
	for x := 6; x < 10; x++ {
		img.SetNRGBA(x, 7, fg)
	}
	for y := 7; y < 11; y++ {
		img.SetNRGBA(6, y, fg)
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

type commandBrowserOpener struct{}

func (commandBrowserOpener) Open(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start browser command: %w", err)
	}
	return nil
}
