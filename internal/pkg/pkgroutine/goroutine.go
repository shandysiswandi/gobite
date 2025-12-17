package pkgroutine

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"sync"
)

// DefaultMaxGoroutine is used when NewManager receives a non-positive limit.
const DefaultMaxGoroutine int = 10

// Manager runs functions in goroutines with a configurable concurrency limit.
//
// It collects errors returned by tasks and can be waited on using Wait.
type Manager struct {
	mu   sync.Mutex
	errs []error
	wg   *sync.WaitGroup
	sema chan struct{}
}

// NewManager creates a new Manager with the provided maximum concurrency.
func NewManager(maxGoroutine int) *Manager {
	if maxGoroutine < 1 {
		maxGoroutine = DefaultMaxGoroutine
	}

	return &Manager{
		wg:   &sync.WaitGroup{},
		sema: make(chan struct{}, maxGoroutine), // Semaphore to limit goroutines
	}
}

// Go schedules a function to run in a goroutine if capacity is available.
//
// If the manager is already at its concurrency limit, the function is not run
// and a warning is logged.
func (g *Manager) Go(pCtx context.Context, f func(ctx context.Context) error) {
	select {
	case g.sema <- struct{}{}: // Acquire a semaphore slot
		g.wg.Go(func() {
			defer func() {
				<-g.sema // Release semaphore slot

				if rvr := recover(); rvr != nil {
					stack := debug.Stack()
					slog.ErrorContext(pCtx, "panic occurred in goroutine", "stack", string(stack))
				}
			}()

			select {
			case <-pCtx.Done():
				slog.WarnContext(pCtx, "goroutine canceled", "because", pCtx.Err())
			default:
				if err := f(pCtx); err != nil {
					g.mu.Lock()
					g.errs = append(g.errs, err)
					g.mu.Unlock()
				}
			}
		})

	default:
		slog.WarnContext(pCtx, "Maximum goroutine limit reached, failed to start new goroutine")
	}
}

// Wait blocks until all scheduled goroutines finish and returns any collected errors.
func (g *Manager) Wait() error {
	g.wg.Wait()

	return errors.Join(g.errs...)
}
