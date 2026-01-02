package pgxcasbin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/lo"
)

const (
	defaultTableName = "casbin_rule"

	insertRow     = "insert into %[1]s (ptype, %[2]s) values ($1, %[3]s) on conflict (ptype, %[2]s) do nothing"
	updateRow     = "update %[1]s set %[2]s where ptype = $1 and %[3]s"
	deleteAll     = "truncate table %[1]s restart identity"
	deleteRow     = "delete from %[1]s where ptype = $1 and %[2]s"
	deleteByPType = "delete from %[1]s where ptype = $1"
	selectSQL     = "select ptype, %[2]s from %[1]s"
)

type store struct {
	db                  Commander
	tableName           string
	noRowsAffectedError error
}

// Commander defines the pgx operations required by the adapter store.
type Commander interface {
	Begin(context.Context) (pgx.Tx, error)
	SendBatch(context.Context, *pgx.Batch) pgx.BatchResults
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func newStore(db Commander) *store {
	return &store{db: db, tableName: defaultTableName, noRowsAffectedError: nil}
}

func (s *store) setTableName(tableName string) {
	s.tableName = lo.SnakeCase(tableName)
}

func (s *store) setNoRowsAffectedError(err error) {
	s.noRowsAffectedError = err
}

func (s *store) insertRow(ctx context.Context, ptype string, args ...string) error {
	normalized, err := normalizeRule(args)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(insertRow, s.tableName, strings.Join(lo.Times(6, func(i int) string {
		return "v" + strconv.Itoa(i)
	}), ","), strings.Join(lo.Times(6, func(i int) string {
		return "$" + strconv.Itoa(i+2)
	}), ","))

	tag, err := s.db.Exec(ctx, sql, lo.ToAnySlice(genRule(ptype, normalized))...)
	if err != nil {
		return errors.Join(ErrInsertRow, err)
	}
	if tag.RowsAffected() == 0 {
		return s.noRowsAffectedError
	}
	return nil
}

func (s *store) selectAll(ctx context.Context) ([][]string, error) {
	return s.selectWhere(ctx, "", 0)
}

func (s *store) selectWhere(ctx context.Context, ptype string, startIdx int, args ...string) ([][]string, error) {
	if len(args) > 6-startIdx {
		return nil, fmt.Errorf("%w: %d > %d", ErrArgsTooLong, len(args), 6-startIdx)
	}

	sqlQuery := fmt.Sprintf(selectSQL, s.tableName, strings.Join(lo.Times(6, func(i int) string {
		return "v" + strconv.Itoa(i)
	}), ","), ptype)

	conditions := make([]string, 0, 1+len(args))
	argsList := make([]any, 0, 1+len(args))
	if lo.IsNotEmpty(ptype) {
		conditions = append(conditions, "ptype = $1")
		argsList = append(argsList, ptype)
	}
	for i, arg := range args {
		if lo.IsEmpty(arg) {
			continue
		}
		conditions = append(conditions, "v"+strconv.Itoa(i+startIdx)+" = $"+strconv.Itoa(len(argsList)+1))
		argsList = append(argsList, arg)
	}
	if len(conditions) > 0 {
		sqlQuery += " where " + strings.Join(conditions, " and ")
	}

	rows, err := s.db.Query(ctx, sqlQuery, argsList...)
	if err != nil {
		return nil, errors.Join(ErrSelectWhere, err)
	}
	defer rows.Close()

	var result [][]string
	for rows.Next() {
		row := make([]sql.NullString, 6+1)
		scanArgs := make([]any, len(row))
		for i := range row {
			scanArgs[i] = &row[i]
		}
		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, errors.Join(ErrScanRow, err)
		}
		converted := make([]string, len(row))
		for i := range row {
			if row[i].Valid {
				converted[i] = row[i].String
			} else {
				converted[i] = ""
			}
		}
		result = append(result, trimTrailingEmpty(converted))
	}
	return result, nil
}

func (s *store) updateRow(ctx context.Context, ptype string, old, updated []string) error {
	normalizedOld, err := normalizeRule(old)
	if err != nil {
		return err
	}
	normalizedUpdated, err := normalizeRule(updated)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(updateRow, s.tableName, strings.Join(lo.Times(6, func(i int) string {
		return "v" + strconv.Itoa(i) + " = $" + strconv.Itoa(i+2)
	}), ", "), strings.Join(lo.Times(6, func(i int) string {
		return "v" + strconv.Itoa(i) + " = $" + strconv.Itoa(i+6+2)
	}), " and "))

	normalizedUpdated = append(normalizedUpdated, normalizedOld...)
	tag, err := s.db.Exec(ctx, sql, lo.ToAnySlice(genRule(ptype, normalizedUpdated))...)
	if err != nil {
		return errors.Join(ErrUpdateRow, err)
	}
	if tag.RowsAffected() == 0 {
		return s.noRowsAffectedError
	}
	return nil
}

func (s *store) deleteRow(ctx context.Context, ptype string, args ...string) error {
	normalized, err := normalizeRule(args)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(deleteRow, s.tableName, strings.Join(lo.Times(6, func(i int) string {
		return "v" + strconv.Itoa(i) + " = $" + strconv.Itoa(i+2)
	}), " and "))

	tag, err := s.db.Exec(ctx, sql, lo.ToAnySlice(genRule(ptype, normalized))...)
	if err != nil {
		return errors.Join(ErrDeleteRow, err)
	}
	if tag.RowsAffected() == 0 {
		return s.noRowsAffectedError
	}
	return nil
}

func (s *store) deleteWhere(ctx context.Context, ptype string, startIdx int, args ...string) error {
	if ptype == "" {
		return ErrEmptyPtype
	}

	if len(args) > 6-startIdx {
		return fmt.Errorf("%w: %d > %d", ErrArgsTooLong, len(args), 6-startIdx)
	}

	sql := fmt.Sprintf(deleteByPType, s.tableName)

	conditions := make([]string, 0, len(args))
	argsList := make([]any, 0, 1+len(args))
	argsList = append(argsList, ptype)
	for i, arg := range args {
		if lo.IsEmpty(arg) {
			continue
		}
		conditions = append(conditions, "v"+strconv.Itoa(i+startIdx)+" = $"+strconv.Itoa(len(argsList)+1))
		argsList = append(argsList, arg)
	}

	if len(conditions) > 0 {
		sql += " and " + strings.Join(conditions, " and ")
	}

	tag, err := s.db.Exec(ctx, sql, argsList...)
	if err != nil {
		return errors.Join(ErrDeleteWhere, err)
	}
	if tag.RowsAffected() == 0 {
		return s.noRowsAffectedError
	}
	return nil
}

func (s *store) deleteAndInsertAll(ctx context.Context, rules [][]string) (err error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errors.Join(ErrBeginTx, err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				err = errors.Join(err, ErrRollbackTx, rbErr)
			}
		}
	}()

	_, err = tx.Exec(ctx, fmt.Sprintf(deleteAll, s.tableName))
	if err != nil {
		return errors.Join(ErrDeleteAll, err)
	}

	if err = s.insertRulesBatch(ctx, tx, rules); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Join(ErrCommitTx, err)
	}
	return nil
}

func (s *store) batchInsert(ctx context.Context, ptype string, rules [][]string) error {
	if len(rules) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, rule := range rules {
		normalized, err := normalizeRule(rule)
		if err != nil {
			return err
		}
		sql := fmt.Sprintf(insertRow, s.tableName, strings.Join(lo.Times(6, func(i int) string {
			return "v" + strconv.Itoa(i)
		}), ","), strings.Join(lo.Times(6, func(i int) string {
			return "$" + strconv.Itoa(i+2)
		}), ","))
		batch.Queue(sql, lo.ToAnySlice(genRule(ptype, normalized))...)
	}

	br := s.db.SendBatch(ctx, batch)

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			closeErr := closeBatchResults(br)
			return errors.Join(ErrBatchExec, err, closeErr)
		}
	}
	if err := closeBatchResults(br); err != nil {
		return err
	}
	return nil
}

func (s *store) batchDelete(ctx context.Context, ptype string, rules [][]string) error {
	if len(rules) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, rule := range rules {
		normalized, err := normalizeRule(rule)
		if err != nil {
			return err
		}
		sql := fmt.Sprintf(deleteRow, s.tableName, strings.Join(lo.Times(6, func(i int) string {
			return "v" + strconv.Itoa(i) + " = $" + strconv.Itoa(i+2)
		}), " and "))
		batch.Queue(sql, lo.ToAnySlice(genRule(ptype, normalized))...)
	}

	br := s.db.SendBatch(ctx, batch)

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			closeErr := closeBatchResults(br)
			return errors.Join(ErrBatchExec, err, closeErr)
		}
	}
	if err := closeBatchResults(br); err != nil {
		return err
	}
	return nil
}

func (s *store) batchUpdate(ctx context.Context, ptype string, oldRules, newRules [][]string) error {
	if len(oldRules) == 0 || len(newRules) == 0 {
		return nil
	}
	if len(oldRules) != len(newRules) {
		return fmt.Errorf("%w: %d vs %d", ErrRulesMismatch, len(oldRules), len(newRules))
	}

	batch := &pgx.Batch{}
	for i := 0; i < len(oldRules); i++ {
		normalizedOld, err := normalizeRule(oldRules[i])
		if err != nil {
			return err
		}
		normalizedNew, err := normalizeRule(newRules[i])
		if err != nil {
			return err
		}
		sql := fmt.Sprintf(updateRow, s.tableName, strings.Join(lo.Times(6, func(i int) string {
			return "v" + strconv.Itoa(i) + " = $" + strconv.Itoa(i+2)
		}), ", "), strings.Join(lo.Times(6, func(i int) string {
			return "v" + strconv.Itoa(i) + " = $" + strconv.Itoa(i+6+2)
		}), " and "))

		normalizedNew = append(normalizedNew, normalizedOld...)
		batch.Queue(sql, lo.ToAnySlice(genRule(ptype, normalizedNew))...)
	}

	br := s.db.SendBatch(ctx, batch)

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			closeErr := closeBatchResults(br)
			return errors.Join(ErrBatchExec, err, closeErr)
		}
	}
	if err := closeBatchResults(br); err != nil {
		return err
	}
	return nil
}

func closeBatchResults(br pgx.BatchResults) error {
	if err := br.Close(); err != nil {
		return errors.Join(ErrBatchClose, err)
	}
	return nil
}

func (s *store) insertRulesBatch(ctx context.Context, tx pgx.Tx, rules [][]string) error {
	if len(rules) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, rule := range rules {
		normalized, err := normalizeRuleRow(rule)
		if err != nil {
			return err
		}
		sql := fmt.Sprintf(insertRow, s.tableName, strings.Join(lo.Times(6, func(i int) string {
			return "v" + strconv.Itoa(i)
		}), ","), strings.Join(lo.Times(6, func(i int) string {
			return "$" + strconv.Itoa(i+2)
		}), ","))
		batch.Queue(sql, lo.ToAnySlice(normalized)...)
	}

	if batch.Len() == 0 {
		return nil
	}

	br := tx.SendBatch(ctx, batch)

	for i := 0; i < batch.Len(); i++ {
		tag, err := br.Exec()
		if err != nil {
			closeErr := closeBatchResults(br)
			return errors.Join(ErrBatchExec, err, closeErr)
		}
		if tag.RowsAffected() == 0 && s.noRowsAffectedError != nil {
			closeErr := closeBatchResults(br)
			return errors.Join(s.noRowsAffectedError, closeErr)
		}
	}

	if err := closeBatchResults(br); err != nil {
		return err
	}
	return nil
}

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
