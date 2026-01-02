package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type DB struct {
	conn  *pgxpool.Pool
	query *sqlc.Queries
	ins   instrument.Instrumentation
}

func NewDB(conn *pgxpool.Pool, ins instrument.Instrumentation) *DB {
	return &DB{
		conn:  conn,
		query: sqlc.New(conn),
		ins:   ins,
	}
}

// - 23505 unique violation → maybe goerror.ErrConflict
// - 23503 foreign_key_violation → maybe goerror.ErrNotFound or a specific “invalid reference”
// - 23502 not_null_violation → goerror.ErrInvalid / validation
// - 23514 check_violation → goerror.ErrInvalid
// - 40001 serialization_failure → retryable error
// - 40P01 deadlock_detected → retryable error
func (s *DB) mapError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return goerror.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return goerror.ErrConflict
	}

	return err
}

func (s *DB) startSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return s.ins.Tracer("identity.outbound.db").Start(ctx, name)
}

func (s *DB) endSpan(span trace.Span, err error) {
	if err != nil && !errors.Is(err, goerror.ErrNotFound) && !errors.Is(err, goerror.ErrConflict) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

func (s *DB) DeleteChallenge(ctx context.Context, id int64) (err error) {
	ctx, span := s.startSpan(ctx, "DeleteChallenge")
	defer func() { s.endSpan(span, err) }()

	err = s.mapError(s.query.DeleteIdentityChallengeByID(ctx, id))
	return err
}
