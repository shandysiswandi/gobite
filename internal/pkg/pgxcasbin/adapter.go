package pgxcasbin

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	"github.com/samber/lo"
	"go.uber.org/atomic"
)

// Adapter stores and retrieves Casbin policies using pgx.
type Adapter struct {
	store  *store
	filter *atomic.Bool
}

var (
	_ persist.Adapter                 = (*Adapter)(nil)
	_ persist.ContextAdapter          = (*Adapter)(nil)
	_ persist.FilteredAdapter         = (*Adapter)(nil)
	_ persist.ContextFilteredAdapter  = (*Adapter)(nil)
	_ persist.BatchAdapter            = (*Adapter)(nil)
	_ persist.ContextBatchAdapter     = (*Adapter)(nil)
	_ persist.UpdatableAdapter        = (*Adapter)(nil)
	_ persist.ContextUpdatableAdapter = (*Adapter)(nil)
)

// Option configures a pgxcasbin Adapter.
type Option func(*Adapter)

// WithTableName overrides the default Casbin rule table name.
func WithTableName(tableName string) Option {
	return func(a *Adapter) {
		a.store.setTableName(tableName)
	}
}

// WithNoRowsAffectedError configures the error used when no rows are affected.
func WithNoRowsAffectedError(err error) Option {
	return func(a *Adapter) {
		a.store.setNoRowsAffectedError(err)
	}
}

// NewAdapter creates a pgx-backed Casbin adapter and ensures the table exists.
func NewAdapter(ctx context.Context, db interface {
	driver.Pinger
	Commander
}, opts ...Option) (*Adapter, error) {
	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	adapter := &Adapter{
		store:  newStore(db),
		filter: atomic.NewBool(false),
	}
	for _, opt := range opts {
		opt(adapter)
	}

	return adapter, nil
}

// LoadPolicyCtx loads all policies into the model.
func (a *Adapter) LoadPolicyCtx(ctx context.Context, model model.Model) error {
	a.filter.Store(false)
	lines, err := a.store.selectAll(ctx)
	if err != nil {
		return err
	}
	return loadLines(model, lines)
}

// SavePolicyCtx persists all policies from the model.
func (a *Adapter) SavePolicyCtx(ctx context.Context, model model.Model) error {
	return a.store.deleteAndInsertAll(ctx, collectRules(model))
}

// AddPolicyCtx adds a single policy rule.
func (a *Adapter) AddPolicyCtx(ctx context.Context, sec string, ptype string, rule []string) error {
	return a.store.insertRow(ctx, ptype, rule...)
}

// RemovePolicyCtx removes a single policy rule.
func (a *Adapter) RemovePolicyCtx(ctx context.Context, sec string, ptype string, rule []string) error {
	return a.store.deleteRow(ctx, ptype, rule...)
}

// RemoveFilteredPolicyCtx removes policy rules matching the filter.
func (a *Adapter) RemoveFilteredPolicyCtx(ctx context.Context, sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return a.store.deleteWhere(ctx, ptype, fieldIndex, fieldValues...)
}

// UpdatePolicyCtx updates a single policy rule.
func (a *Adapter) UpdatePolicyCtx(ctx context.Context, sec string, ptype string, oldRule, newRule []string) error {
	return a.store.updateRow(ctx, ptype, oldRule, newRule)
}

// UpdatePoliciesCtx updates multiple policy rules.
func (a *Adapter) UpdatePoliciesCtx(ctx context.Context, sec string, ptype string, oldRules, newRules [][]string) error {
	return a.store.batchUpdate(ctx, ptype, oldRules, newRules)
}

// UpdateFilteredPoliciesCtx replaces filtered policies with new rules.
func (a *Adapter) UpdateFilteredPoliciesCtx(ctx context.Context, sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	oldRows, err := a.store.selectWhere(ctx, ptype, fieldIndex, fieldValues...)
	if err != nil {
		return nil, err
	}
	if err := a.store.deleteWhere(ctx, ptype, fieldIndex, fieldValues...); err != nil {
		return nil, err
	}
	if err := a.store.batchInsert(ctx, ptype, newRules); err != nil {
		return nil, err
	}

	oldRules := make([][]string, 0, len(oldRows))
	for _, row := range oldRows {
		if len(row) == 0 {
			continue
		}
		oldRules = append(oldRules, row[1:])
	}
	return oldRules, nil
}

// LoadFilteredPolicyCtx loads policies matching the filter into the model.
func (a *Adapter) LoadFilteredPolicyCtx(ctx context.Context, model model.Model, filter interface{}) error {
	if lo.IsNil(filter) {
		return a.LoadPolicyCtx(ctx, model)
	}
	a.filter.Store(true)
	ft, ok := filter.(map[string][][]string)
	if !ok {
		return fmt.Errorf("%w: got %T. Expected format: map[role][][]fieldValues, "+
			"where role is the key (like 'p', 'g', 'g2') and fieldValues is a 2D slice of strings representing multiple conditions with OR relationship between them. "+
			"For unused attributes in conditions, use an empty string", ErrInvalidFilterType, filter)
	}
	var lines [][]string
	for k, v := range ft {
		for _, vv := range v {
			tmp, err := a.store.selectWhere(ctx, k, 0, vv...)
			if err != nil {
				return err
			}
			lines = append(lines, tmp...)
		}
	}
	lines = lo.UniqBy(lines, func(line []string) string {
		return strings.Join(line, ",")
	})
	if len(lines) == 0 {
		return nil
	}
	return loadLines(model, lines)
}

// IsFilteredCtx reports whether the last load used a filter.
func (a *Adapter) IsFilteredCtx(ctx context.Context) bool {
	return a.filter.Load()
}

// LoadPolicy loads all policies into the model.
func (a *Adapter) LoadPolicy(model model.Model) error {
	return a.LoadPolicyCtx(context.Background(), model)
}

// SavePolicy persists all policies from the model.
func (a *Adapter) SavePolicy(model model.Model) error {
	return a.SavePolicyCtx(context.Background(), model)
}

// AddPolicy adds a single policy rule.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return a.AddPolicyCtx(context.Background(), sec, ptype, rule)
}

// RemovePolicy removes a single policy rule.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return a.RemovePolicyCtx(context.Background(), sec, ptype, rule)
}

// RemoveFilteredPolicy removes policy rules matching the filter.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return a.RemoveFilteredPolicyCtx(context.Background(), sec, ptype, fieldIndex, fieldValues...)
}

// UpdatePolicy updates a single policy rule.
func (a *Adapter) UpdatePolicy(sec string, ptype string, oldRule, newRule []string) error {
	return a.UpdatePolicyCtx(context.Background(), sec, ptype, oldRule, newRule)
}

// UpdatePolicies updates multiple policy rules.
func (a *Adapter) UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	return a.UpdatePoliciesCtx(context.Background(), sec, ptype, oldRules, newRules)
}

// UpdateFilteredPolicies replaces filtered policies with new rules.
func (a *Adapter) UpdateFilteredPolicies(sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	return a.UpdateFilteredPoliciesCtx(context.Background(), sec, ptype, newRules, fieldIndex, fieldValues...)
}

// AddPolicies adds multiple policy rules.
func (a *Adapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	return a.AddPoliciesCtx(context.Background(), sec, ptype, rules)
}

// RemovePolicies removes multiple policy rules.
func (a *Adapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	return a.RemovePoliciesCtx(context.Background(), sec, ptype, rules)
}

// LoadFilteredPolicy loads policies matching the filter into the model.
func (a *Adapter) LoadFilteredPolicy(model model.Model, filter interface{}) error {
	return a.LoadFilteredPolicyCtx(context.Background(), model, filter)
}

// IsFiltered reports whether the last load used a filter.
func (a *Adapter) IsFiltered() bool {
	return a.IsFilteredCtx(context.Background())
}

// AddPoliciesCtx adds multiple policy rules.
func (a *Adapter) AddPoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) error {
	return a.store.batchInsert(ctx, ptype, rules)
}

// RemovePoliciesCtx removes multiple policy rules.
func (a *Adapter) RemovePoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) error {
	return a.store.batchDelete(ctx, ptype, rules)
}

func collectRules(model model.Model) [][]string {
	var rules [][]string

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			rules = append(rules, genRule(ptype, rule))
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			rules = append(rules, genRule(ptype, rule))
		}
	}

	return rules
}

func loadLines(model model.Model, lines [][]string) error {
	for _, line := range lines {
		if err := persist.LoadPolicyArray(line, model); err != nil {
			return err
		}
	}
	return nil
}
