package pgxcasbin

import "errors"

var (
	// ErrInvalidFilterType indicates the filter value is not supported.
	ErrInvalidFilterType = errors.New("invalid filter type")
	// ErrBatchExec indicates a batch execution failure.
	ErrBatchExec = errors.New("failed to execute batch")
	// ErrBatchClose indicates a batch close failure.
	ErrBatchClose = errors.New("failed to close batch")
	// ErrInsertRow indicates a row insert failure.
	ErrInsertRow = errors.New("failed to insert row")
	// ErrArgsTooLong indicates the provided args exceed the field count.
	ErrArgsTooLong = errors.New("args length exceeds field count")
	// ErrSelectWhere indicates a filtered select failure.
	ErrSelectWhere = errors.New("failed to select where")
	// ErrScanRow indicates a row scan failure.
	ErrScanRow = errors.New("failed to scan row")
	// ErrUpdateRow indicates a row update failure.
	ErrUpdateRow = errors.New("failed to update row")
	// ErrDeleteRow indicates a row delete failure.
	ErrDeleteRow = errors.New("failed to delete row")
	// ErrEmptyPtype indicates a missing policy type.
	ErrEmptyPtype = errors.New("ptype is empty")
	// ErrDeleteWhere indicates a filtered delete failure.
	ErrDeleteWhere = errors.New("failed to delete where")
	// ErrBeginTx indicates a transaction begin failure.
	ErrBeginTx = errors.New("failed to begin transaction")
	// ErrDeleteAll indicates a truncate failure.
	ErrDeleteAll = errors.New("failed to delete all rows")
	// ErrCommitTx indicates a transaction commit failure.
	ErrCommitTx = errors.New("failed to commit transaction")
	// ErrRollbackTx indicates a transaction rollback failure.
	ErrRollbackTx = errors.New("failed to rollback transaction")
	// ErrRulesMismatch indicates mismatched rule counts.
	ErrRulesMismatch = errors.New("oldRules and newRules length mismatch")
	// ErrRuleTooLong indicates a rule exceeds the field count.
	ErrRuleTooLong = errors.New("rule length exceeds field count")
	// ErrRuleEmpty indicates an empty rule payload.
	ErrRuleEmpty = errors.New("rule is empty")
	// ErrNewPool indicates a pgx pool creation failure.
	ErrNewPool = errors.New("failed to create pgx pool")
	// ErrPingPool indicates a pool ping failure.
	ErrPingPool = errors.New("failed to ping pool")
	// ErrUnknownUpdateType indicates an unsupported update message.
	ErrUnknownUpdateType = errors.New("unknown update type")
	// ErrMarshalMessage indicates a message marshal failure.
	ErrMarshalMessage = errors.New("failed to marshal message")
	// ErrNotifyMessage indicates a notify failure.
	ErrNotifyMessage = errors.New("failed to notify")
	// ErrAcquireConn indicates a database connection acquisition failure.
	ErrAcquireConn = errors.New("failed to acquire psql connection")
	// ErrListenChannel indicates a listen channel failure.
	ErrListenChannel = errors.New("failed to listen channel")
	// ErrWaitNotification indicates a notification wait failure.
	ErrWaitNotification = errors.New("failed to wait for notification")
)
