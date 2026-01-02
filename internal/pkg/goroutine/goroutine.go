package goroutine

import (
	"context"
	"errors"
	"log/slog"
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/shandysiswandi/gobite/internal/pkg/stacktrace"
)

// DefaultMaxGoroutine is used when NewManager receives a non-positive limit.
const DefaultMaxGoroutine int = 100

// Manager runs functions in goroutines with a configurable concurrency limit.
//
// It collects errors returned by tasks and can be waited on using Wait.
type Manager struct {
	mu      sync.Mutex
	errs    []error
	wg      *sync.WaitGroup
	sema    chan struct{}
	stateMu sync.RWMutex
	closed  bool
}

// NewManager creates a new Manager with the provided maximum concurrency.
func NewManager(maxGoroutine int) *Manager {
	if maxGoroutine < 1 {
		maxGoroutine = runtime.NumCPU() * DefaultMaxGoroutine
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
	if g == nil {
		return
	}

	g.stateMu.RLock()
	if g.closed {
		g.stateMu.RUnlock()
		slog.WarnContext(pCtx, "goroutine manager is closed, skipping new goroutine")
		return
	}

	select {
	case g.sema <- struct{}{}: // Acquire a semaphore slot
		g.wg.Go(func() {
			g.stateMu.RUnlock()
			defer func() {
				<-g.sema // Release semaphore slot

				if rvr := recover(); rvr != nil {
					stack := debug.Stack()
					paths := stacktrace.InternalPaths(stack)
					if len(paths) == 0 {
						slog.ErrorContext(pCtx, "panic occurred in goroutine", "stack", string(stack))
					} else {
						slog.ErrorContext(pCtx, "panic occurred in goroutine", "stack", paths)
					}
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
		g.stateMu.RUnlock()
		slog.WarnContext(pCtx, "Maximum goroutine limit reached, failed to start new goroutine")
	}
}

// Wait blocks until all scheduled goroutines finish and returns any collected errors.
func (g *Manager) Wait() error {
	if g == nil {
		return nil
	}

	g.stateMu.Lock()
	if !g.closed {
		g.closed = true
	}
	g.stateMu.Unlock()

	g.wg.Wait()

	return errors.Join(g.errs...)
}
