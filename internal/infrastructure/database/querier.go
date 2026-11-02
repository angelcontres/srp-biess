package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier abstrae la ejecución de sentencias SQL en pgx/v5.
// Permite que un repositorio opere indistintamente sobre el pool de conexiones (*pgxpool.Pool)
// o dentro de una transacción atómica activa (pgx.Tx) (ADR-004, SRP-004).
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
