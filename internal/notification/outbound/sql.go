package outbound

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

type SQL struct {
	conn  *pgxpool.Pool
	query *sqlc.Queries
}

func NewSQL(conn *pgxpool.Pool) *SQL {
	return &SQL{
		conn:  conn,
		query: sqlc.New(conn),
	}
}
