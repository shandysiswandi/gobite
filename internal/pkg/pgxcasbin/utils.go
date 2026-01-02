package pgxcasbin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func genRule(ptype string, rule []string) []string {
	result := make([]string, 1+len(rule))
	result[0] = ptype
	copy(result[1:], rule)
	return result
}

func normalizeRule(rule []string) ([]string, error) {
	if len(rule) > 6 {
		return nil, fmt.Errorf("%w: %d > %d", ErrRuleTooLong, len(rule), 6)
	}
	normalized := make([]string, 6)
	copy(normalized, rule)
	return normalized, nil
}

func normalizeRuleRow(rule []string) ([]string, error) {
	if len(rule) == 0 {
		return nil, ErrRuleEmpty
	}
	normalized, err := normalizeRule(rule[1:])
	if err != nil {
		return nil, err
	}
	return genRule(rule[0], normalized), nil
}

func trimTrailingEmpty(rule []string) []string {
	last := len(rule) - 1
	for last >= 0 && rule[last] == "" {
		last--
	}
	return rule[:last+1]
}

// Commander defines the pgx operations required by the adapter store.
type Commander interface {
	Begin(context.Context) (pgx.Tx, error)
	SendBatch(context.Context, *pgx.Batch) pgx.BatchResults
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
