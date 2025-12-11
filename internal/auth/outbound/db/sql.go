package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgsql"
)

type SQL struct {
	conn  *pgxpool.Pool
	query *pkgsql.Queries
}

func NewSQL(conn *pgxpool.Pool) *SQL {
	return &SQL{
		conn:  conn,
		query: pkgsql.New(conn),
	}
}
