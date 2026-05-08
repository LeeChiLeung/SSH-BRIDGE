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
	server := &http.Server{Addr: s.addr, Handler: s.router.Router()}
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
