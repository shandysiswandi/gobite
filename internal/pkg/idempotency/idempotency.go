package idempotency

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrAlreadyInProgress = errors.New("operation already in progress")
	ErrAlreadyCompleted  = errors.New("operation already completed")
	ErrAlreadyFailed     = errors.New("operation already failed")
	ErrInvalidState      = errors.New("invalid state")
)

type State string

const (
	StateNone       State = "none"        // operation can proceed
	StateInProgress State = "in_progress" // operation already in progress
	StateCompleted  State = "completed"   // operation already completed
	StateFailed     State = "failed"      // previously operation failed
	StateError      State = "error"       // this operation error
)

func (s State) String() string {
	return string(s)
}

type Idempotency interface {
	Acquire(ctx context.Context, key string, lockDuration time.Duration) (State, error)
	MarkCompleted(ctx context.Context, key string, ttl time.Duration) error
	MarkFailed(ctx context.Context, key string, ttl time.Duration) error
	Exec(ctx context.Context, key string, fn func(context.Context) error, opts ...Option) error
}

type StateTracker struct {
	client *redis.Client
	prefix string
}

func New(client *redis.Client) *StateTracker {
	return &StateTracker{
		client: client,
		prefix: "idempotency:",
	}
}

const (
	defaultLockDuration = time.Minute
	defaultStateTTL     = time.Minute
)

type Option func(*execOptions)

type execOptions struct {
	lockDuration time.Duration
	stateTTL     time.Duration
}

func WithLockDuration(lockDuration time.Duration) Option {
	return func(o *execOptions) {
		o.lockDuration = lockDuration
	}
}

func WithStateTTL(stateTTL time.Duration) Option {
	return func(o *execOptions) {
		o.stateTTL = stateTTL
	}
}

// Acquire tries to start an operation
func (s *StateTracker) Acquire(ctx context.Context, key string, lockDuration time.Duration) (State, error) {
	fk := s.prefix + key

	acquired, err := s.client.SetNX(ctx, fk, StateInProgress.String(), lockDuration).Result()
	if err != nil {
		return StateError, err
	}
	if acquired {
		return StateNone, nil
	}

	result, err := s.client.Get(ctx, fk).Result()
	if errors.Is(err, redis.Nil) {
		acquired, err = s.client.SetNX(ctx, fk, StateInProgress.String(), lockDuration).Result()
		if err != nil {
			return StateError, err
		}
		if acquired {
			return StateNone, nil
		}
		return StateError, ErrInvalidState
	}
	if err != nil {
		return StateError, err
	}

	switch result {
	case StateInProgress.String():
		return StateInProgress, nil
	case StateCompleted.String():
		return StateCompleted, nil
	case StateFailed.String():
		return StateFailed, nil
	default:
		return StateError, ErrInvalidState
	}
}

func (s *StateTracker) MarkCompleted(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.Set(ctx, s.prefix+key, StateCompleted.String(), ttl).Err()
}

func (s *StateTracker) MarkFailed(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.Set(ctx, s.prefix+key, StateFailed.String(), ttl).Err()
}

func (s *StateTracker) Exec(ctx context.Context, key string, fn func(context.Context) error, opts ...Option) error {
	execOpt := &execOptions{
		lockDuration: defaultLockDuration,
		stateTTL:     defaultStateTTL,
	}
	for _, opt := range opts {
		opt(execOpt)
	}
	if execOpt.lockDuration <= 0 {
		execOpt.lockDuration = defaultLockDuration
	}
	if execOpt.stateTTL <= 0 {
		execOpt.stateTTL = defaultStateTTL
	}

	state, err := s.Acquire(ctx, key, execOpt.lockDuration)
	if err != nil {
		return err
	}

	switch state {
	case StateInProgress:
		return ErrAlreadyInProgress
	case StateCompleted:
		return ErrAlreadyCompleted
	case StateFailed:
		return ErrAlreadyFailed
	}

	if err := fn(ctx); err != nil {
		if markErr := s.MarkFailed(ctx, key, execOpt.stateTTL); markErr != nil {
			return markErr
		}
		return err
	}

	if err := s.MarkCompleted(ctx, key, execOpt.stateTTL); err != nil {
		return err
	}

	return nil
}
