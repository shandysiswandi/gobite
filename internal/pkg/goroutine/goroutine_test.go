package goroutine

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestManagerGoAndWait(t *testing.T) {
	cases := []struct {
		name        string
		max         int
		tasks       int
		wantRuns    int32
		wantErr     bool
		useBlocking bool
	}{
		{name: "single_task", max: 2, tasks: 1, wantRuns: 1, wantErr: true},
		{name: "limit_enforced", max: 1, tasks: 2, wantRuns: 1, wantErr: false, useBlocking: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mgr := NewManager(tc.max)
			ctx := context.Background()

			var runs atomic.Int32
			started := make(chan struct{})
			release := make(chan struct{})

			fn := func(ctx context.Context) error {
				runs.Add(1)
				if tc.useBlocking {
					close(started)
					select {
					case <-release:
					case <-time.After(time.Second):
						return errors.New("timeout")
					}
				}
				if tc.wantErr {
					return errors.New("fail")
				}
				return nil
			}

			mgr.Go(ctx, fn)
			if tc.tasks > 1 {
				if tc.useBlocking {
					<-started
					mgr.Go(ctx, fn)
					close(release)
				} else {
					mgr.Go(ctx, fn)
				}
			}

			err := mgr.Wait()
			if runs.Load() != tc.wantRuns {
				t.Fatalf("runs = %d, want %d", runs.Load(), tc.wantRuns)
			}
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
