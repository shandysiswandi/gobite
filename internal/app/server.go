package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Start launches the HTTP server and returns a channel closed on shutdown.
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
		slog.Info("sse server listening", "address", a.sseServer.Addr)

		if err := a.sseServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to listen and serve sse server", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		defer signal.Stop(sigint)

		<-sigint

		if a.cancel != nil {
			a.cancel()
		}

		close(terminateChan)

		slog.Info("application gracefully shutdown")
	}()

	return terminateChan
}

// Serve runs the HTTP server on the provided listener for tests.
func (a *App) Serve(l net.Listener) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- a.httpServer.Serve(l)
		close(errChan)
	}()

	return errChan
}

// Stop gracefully shuts down the server and closes resources.
func (a *App) Stop(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}

	if err := a.httpServer.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to close resources", "name", "HTTP Server", "error", err)
	}
	if err := a.sseServer.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to close resources", "name", "SSE Server", "error", err)
	}

	slog.InfoContext(ctx, "waiting for all goroutine to finish")
	if err := a.goroutine.Wait(); err != nil {
		slog.ErrorContext(ctx, "error from goroutines executions", "error", err)
	}
	slog.InfoContext(ctx, "all goroutines have finished successfully")

	for _, closer := range a.closers {
		if err := closer.fn(ctx); err != nil {
			slog.ErrorContext(ctx, "failed to close resources", "name", closer.name, "error", err)
		}
	}
}
