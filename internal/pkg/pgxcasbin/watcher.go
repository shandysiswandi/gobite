package pgxcasbin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sethvargo/go-retry"
)

// UpdateType represents the type of policy update message.
type UpdateType string

const (
	// Update requests a full policy reload.
	Update UpdateType = "Update"
	// UpdateForAddPolicy adds a single policy rule.
	UpdateForAddPolicy UpdateType = "UpdateForAddPolicy"
	// UpdateForRemovePolicy removes a single policy rule.
	UpdateForRemovePolicy UpdateType = "UpdateForRemovePolicy"
	// UpdateForRemoveFilteredPolicy removes policies by filter.
	UpdateForRemoveFilteredPolicy UpdateType = "UpdateForRemoveFilteredPolicy"
	// UpdateForSavePolicy saves all policies.
	UpdateForSavePolicy UpdateType = "UpdateForSavePolicy"
	// UpdateForAddPolicies adds multiple policy rules.
	UpdateForAddPolicies UpdateType = "UpdateForAddPolicies"
	// UpdateForRemovePolicies removes multiple policy rules.
	UpdateForRemovePolicies UpdateType = "UpdateForRemovePolicies"
	// UpdateForUpdatePolicy updates a single policy rule.
	UpdateForUpdatePolicy UpdateType = "UpdateForUpdatePolicy"
	// UpdateForUpdatePolicies updates multiple policy rules.
	UpdateForUpdatePolicies UpdateType = "UpdateForUpdatePolicies"
)

// Watcher listens for policy changes and triggers callbacks.
type Watcher struct {
	// RWMutex guards watcher state.
	sync.RWMutex

	opt        OptionWatcher
	pool       *pgxpool.Pool
	ownsPool   bool
	callback   func(string)
	cancelFunc func()
}

const defaultChannel = "iam_casbin_psql_watcher"

// OptionWatcher configures a Watcher instance.
type OptionWatcher struct {
	// Channel sets the Postgres listen channel.
	Channel string
	// Verbose enables verbose logging.
	Verbose bool
	// LocalID identifies this watcher instance.
	LocalID string
	// NotifySelf indicates whether self-originated events are emitted.
	NotifySelf bool
}

// GetChannel returns the configured channel name.
func (w *Watcher) GetChannel() string {
	return w.opt.Channel
}

// GetVerbose reports whether verbose logging is enabled.
func (w *Watcher) GetVerbose() bool {
	return w.opt.Verbose
}

// GetLocalID returns the watcher local identifier.
func (w *Watcher) GetLocalID() string {
	return w.opt.LocalID
}

// GetNotifySelf reports whether self notifications are enabled.
func (w *Watcher) GetNotifySelf() bool {
	return w.opt.NotifySelf
}

// MSG describes a policy update message payload.
type MSG struct {
	// Method is the update type.
	Method UpdateType `json:"method"`
	// ID is the sender identifier.
	ID string `json:"id"`
	// Sec is the Casbin section key.
	Sec string `json:"sec,omitempty"`
	// Ptype is the policy type.
	Ptype string `json:"ptype,omitempty"`
	// OldRules holds previous policy rules.
	OldRules [][]string `json:"old_rules,omitempty"`
	// NewRules holds new policy rules.
	NewRules [][]string `json:"new_rules,omitempty"`
	// FieldIndex is the start index for filtered rules.
	FieldIndex int `json:"field_index,omitempty"`
	// FieldValues holds filter values for updates.
	FieldValues []string `json:"field_values,omitempty"`
}

// NewWatcherWithPool creates a Watcher using an existing pgx pool.
func NewWatcherWithPool(ctx context.Context, pool *pgxpool.Pool, opt OptionWatcher) (*Watcher, error) {
	return newWatcherWithPool(ctx, pool, opt, false)
}

func newWatcherWithPool(ctx context.Context, pool *pgxpool.Pool, opt OptionWatcher, ownsPool bool) (*Watcher, error) {
	if err := pool.Ping(ctx); err != nil {
		return nil, errors.Join(ErrPingPool, err)
	}

	if opt.Channel == "" {
		opt.Channel = defaultChannel
	}
	if opt.LocalID == "" {
		opt.LocalID = uuid.New().String()
	}

	listenerCtx, cancel := context.WithCancel(ctx)
	w := &Watcher{
		opt:        opt,
		pool:       pool,
		ownsPool:   ownsPool,
		cancelFunc: cancel,
	}

	go func() {
		b := retry.NewFibonacci(200 * time.Millisecond)

		b = retry.WithCappedDuration((5 * time.Second), b)

		if err := retry.Do(listenerCtx, b, func(ctx context.Context) error {
			if err := w.listenMessage(listenerCtx); errors.Is(err, context.Canceled) {
				slog.Info("pgxcasbin watcher closed")
				return nil // context canceled, exit the loop
			} else if err != nil {
				slog.Error("pgxcasbin failed to listen message", "error", err)

				return retry.RetryableError(err)
			}

			return nil
		}); err != nil {
			slog.Error("pgxcasbin listener stopped with error", "error", err)
		}

		slog.Info("pgxcasbin listener exited")
	}()

	return w, nil
}

// DefaultCallback returns a watcher callback that updates the enforcer.
func DefaultCallback(e casbin.IEnforcer) func(string) {
	return func(s string) {
		var m MSG
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			slog.Error("pgxcasbin unable to unmarshal payload", "payload", s, "error", err)
			return
		}

		var res bool
		var err error
		switch m.Method {
		case Update, UpdateForSavePolicy:
			err = e.LoadPolicy()
			res = true
		case UpdateForAddPolicy:
			if len(m.NewRules) == 0 {
				slog.Warn("pgxcasbin missing new rules for add policy")
				return
			}
			res, err = e.SelfAddPolicy(m.Sec, m.Ptype, m.NewRules[0])
		case UpdateForAddPolicies:
			res, err = e.SelfAddPolicies(m.Sec, m.Ptype, m.NewRules)
		case UpdateForRemovePolicy:
			if len(m.NewRules) == 0 {
				slog.Warn("pgxcasbin missing new rules for remove policy")
				return
			}
			res, err = e.SelfRemovePolicy(m.Sec, m.Ptype, m.NewRules[0])
		case UpdateForRemoveFilteredPolicy:
			res, err = e.SelfRemoveFilteredPolicy(m.Sec, m.Ptype, m.FieldIndex, m.FieldValues...)
		case UpdateForRemovePolicies:
			res, err = e.SelfRemovePolicies(m.Sec, m.Ptype, m.NewRules)
		case UpdateForUpdatePolicy:
			if len(m.OldRules) == 0 || len(m.NewRules) == 0 {
				slog.Warn("pgxcasbin missing old or new rules for update policy")
				return
			}
			res, err = e.SelfUpdatePolicy(m.Sec, m.Ptype, m.OldRules[0], m.NewRules[0])
		case UpdateForUpdatePolicies:
			res, err = e.SelfUpdatePolicies(m.Sec, m.Ptype, m.OldRules, m.NewRules)
		default:
			err = fmt.Errorf("%w: %s", ErrUnknownUpdateType, m.Method)
		}
		if err != nil {
			slog.Error("pgxcasbin failed to update policy", "error", err)
		}
		if !res {
			slog.Warn("pgxcasbin callback update policy failed")
		}
	}
}

// SetUpdateCallback registers the handler invoked on update messages.
func (w *Watcher) SetUpdateCallback(callback func(string)) error {
	w.Lock()
	defer w.Unlock()
	w.callback = callback
	return nil
}

// Update sends a generic update notification.
func (w *Watcher) Update() error {
	return w.notifyMessage(&MSG{
		Method: Update,
		ID:     w.GetLocalID(),
	})
}

// Close stops the watcher and closes its pool if owned.
func (w *Watcher) Close() {
	w.cancelFunc()
	if w.ownsPool && w.pool != nil {
		w.pool.Close()
	}
}

// UpdateForAddPolicy sends a notification for adding a policy rule.
func (w *Watcher) UpdateForAddPolicy(sec, ptype string, params ...string) error {
	return w.notifyMessage(&MSG{
		Method:   UpdateForAddPolicy,
		ID:       w.GetLocalID(),
		Sec:      sec,
		Ptype:    ptype,
		NewRules: [][]string{params},
	})
}

// UpdateForRemovePolicy sends a notification for removing a policy rule.
func (w *Watcher) UpdateForRemovePolicy(sec, ptype string, params ...string) error {
	return w.notifyMessage(&MSG{
		Method:   UpdateForRemovePolicy,
		ID:       w.GetLocalID(),
		Sec:      sec,
		Ptype:    ptype,
		NewRules: [][]string{params},
	})
}

// UpdateForRemoveFilteredPolicy sends a notification for filtered removals.
func (w *Watcher) UpdateForRemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	return w.notifyMessage(&MSG{
		Method:      UpdateForRemoveFilteredPolicy,
		ID:          w.GetLocalID(),
		Sec:         sec,
		Ptype:       ptype,
		FieldIndex:  fieldIndex,
		FieldValues: fieldValues,
	})
}

// UpdateForSavePolicy sends a notification to reload all policies.
func (w *Watcher) UpdateForSavePolicy(model model.Model) error {
	return w.notifyMessage(&MSG{
		Method: UpdateForSavePolicy,
		ID:     w.GetLocalID(),
	})
}

// UpdateForAddPolicies sends a notification for adding multiple policy rules.
func (w *Watcher) UpdateForAddPolicies(sec string, ptype string, rules ...[]string) error {
	return w.notifyMessage(&MSG{
		Method:   UpdateForAddPolicies,
		ID:       w.GetLocalID(),
		Sec:      sec,
		Ptype:    ptype,
		NewRules: rules,
	})
}

// UpdateForRemovePolicies sends a notification for removing multiple rules.
func (w *Watcher) UpdateForRemovePolicies(sec string, ptype string, rules ...[]string) error {
	return w.notifyMessage(&MSG{
		Method:   UpdateForRemovePolicies,
		ID:       w.GetLocalID(),
		Sec:      sec,
		Ptype:    ptype,
		NewRules: rules,
	})
}

// UpdateForUpdatePolicy sends a notification for updating a policy rule.
func (w *Watcher) UpdateForUpdatePolicy(sec string, ptype string, oldRule, newRule []string) error {
	return w.notifyMessage(&MSG{
		Method:   UpdateForUpdatePolicy,
		ID:       w.GetLocalID(),
		Sec:      sec,
		Ptype:    ptype,
		OldRules: [][]string{oldRule},
		NewRules: [][]string{newRule},
	})
}

// UpdateForUpdatePolicies sends a notification for updating multiple rules.
func (w *Watcher) UpdateForUpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	return w.notifyMessage(&MSG{
		Method:   UpdateForUpdatePolicies,
		ID:       w.GetLocalID(),
		Sec:      sec,
		Ptype:    ptype,
		OldRules: oldRules,
		NewRules: newRules,
	})
}

func (w *Watcher) notifyMessage(m *MSG) error {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("%w: %+v", errors.Join(ErrMarshalMessage, err), m)
	}
	cmd := fmt.Sprintf("select pg_notify('%s', $1)", w.GetChannel())

	if _, err := w.pool.Exec(context.Background(), cmd, string(b)); err != nil {
		return fmt.Errorf("%w: %s", errors.Join(ErrNotifyMessage, err), string(b))
	}

	if w.GetVerbose() {
		slog.Info("pgxcasbin send message", "channel", w.GetChannel(), "payload", string(b))
	}

	return nil
}

func (w *Watcher) listenMessage(ctx context.Context) error {
	conn, err := w.pool.Acquire(ctx)
	if err != nil {
		return errors.Join(ErrAcquireConn, err)
	}
	defer conn.Release()

	cmd := fmt.Sprintf("listen %s", w.GetChannel())
	if _, err = conn.Exec(ctx, cmd); err != nil {
		return fmt.Errorf("%w: %s", errors.Join(ErrListenChannel, err), w.GetChannel())
	}

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if errors.Is(err, context.Canceled) {
			return err
		} else if err != nil {
			return errors.Join(ErrWaitNotification, err)
		}

		if w.GetVerbose() {
			slog.Info("pgxcasbin received message", "channel", w.GetChannel(), "local_id", w.GetLocalID(), "payload", notification.Payload)
		}

		var m MSG
		if err := json.Unmarshal([]byte(notification.Payload), &m); err != nil {
			slog.Error("pgxcasbin failed to unmarshal notification", "payload", notification.Payload, "error", err)
			continue
		}

		w.RLock()
		if m.ID != w.GetLocalID() || w.GetNotifySelf() {
			if w.callback != nil {
				w.callback(notification.Payload)
			} else {
				slog.Warn("pgxcasbin callback is not set, skipping update")
			}
		}
		w.RUnlock()
	}
}
