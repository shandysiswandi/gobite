package pkgroutine

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"sync"
)

const DefaultMaxGoroutine int = 10

type Manager struct {
	mu   sync.Mutex
	errs []error
	wg   *sync.WaitGroup
	sema chan struct{}
}

func NewManager(maxGoroutine int) *Manager {
	if maxGoroutine < 1 {
		maxGoroutine = DefaultMaxGoroutine
	}

	return &Manager{
		wg:   &sync.WaitGroup{},
		sema: make(chan struct{}, maxGoroutine), // Semaphore to limit goroutines
	}
}

func (g *Manager) Go(ctx context.Context, f func(c context.Context) error) {
	select {
	case g.sema <- struct{}{}: // Acquire a semaphore slot
		g.wg.Go(func() {
			defer func() {
				<-g.sema // Release semaphore slot

				if rvr := recover(); rvr != nil {
					stack := debug.Stack()
					slog.ErrorContext(ctx, "panic occurred in goroutine", "stack", string(stack))
				}
			}()

			select {
			case <-ctx.Done():
				slog.WarnContext(ctx, "goroutine canceled", "because", ctx.Err())
			default:
				if err := f(ctx); err != nil {
					g.mu.Lock()
					g.errs = append(g.errs, err)
					g.mu.Unlock()
				}
			}
		})

	default:
		slog.WarnContext(ctx, "Maximum goroutine limit reached, failed to start new goroutine")
	}
}

func (g *Manager) Wait() error {
	g.wg.Wait()

	return errors.Join(g.errs...)
}
