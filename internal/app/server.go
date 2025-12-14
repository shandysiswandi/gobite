package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func (a *App) Start() <-chan struct{} {
	terminateChan := make(chan struct{})

	go func() {
		slog.Info("http server listening", "address", a.httpServer.Addr)

		if err := a.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to listen and serve http server", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		<-sigint

		if a.cancel != nil {
			a.cancel()
		}

		terminateChan <- struct{}{}
		close(terminateChan)

		slog.Info("application gracefully shutdown")
	}()

	return terminateChan
}

func (a *App) Stop(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}

	if err := a.httpServer.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to close resources", "name", "HTTP Server", "error", err)
	}

	slog.InfoContext(ctx, "waiting for all goroutine to finish")
	if err := a.goroutine.Wait(); err != nil {
		slog.ErrorContext(ctx, "error from goroutines executions", "error", err)
	}
	slog.InfoContext(ctx, "all goroutines have finished successfully")

	for name, closer := range a.closerFn {
		if name == "HTTP Server" {
			continue
		}
		if err := closer(ctx); err != nil {
			slog.ErrorContext(ctx, "failed to close resources", "name", name, "error", err)
		}
	}
}
