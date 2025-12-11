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

		terminateChan <- struct{}{}
		close(terminateChan)

		slog.Info("application gracefully shutdown")
	}()

	return terminateChan
}

func (a *App) Stop(ctx context.Context) {
	// close resources
	for name, closer := range a.closerFn {
		if err := closer(ctx); err != nil {
			slog.ErrorContext(ctx, "failed to close resources", "name", name, "error", err)
		}
	}

	slog.InfoContext(ctx, "all goroutines have finished successfully")
}
