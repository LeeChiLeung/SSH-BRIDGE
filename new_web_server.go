package main

import (
	"context"
	"net/http"
	"time"

	"timber-ssh-go/ui"
)

type TemplWebServer struct {
	addr   string
	router *ui.App
}

func NewTemplWebServer(addr string, router *ui.App) *TemplWebServer {
	return &TemplWebServer{addr: addr, router: router}
}

func (s *TemplWebServer) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/", s.router.Router())
	mux.HandleFunc("/favicon.ico", serveEmbeddedFavicon)
	mux.HandleFunc("/brand-logo.png", serveEmbeddedBrandLogo)

	server := &http.Server{Addr: s.addr, Handler: mux}
	errCh := make(chan error, 1)

	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func serveEmbeddedFavicon(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(embeddedWindowsTrayIcon)
}

func serveEmbeddedBrandLogo(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(embeddedCustomTrayIcon)
}
