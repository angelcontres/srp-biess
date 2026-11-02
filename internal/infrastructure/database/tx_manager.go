package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"srp-biess/internal/domain"
)

type txContextKey struct{}

// PgxTxManager implementa domain.TxManager coordinando transacciones atómicas con pgx/v5 (ADR-004, SRP-005).
type PgxTxManager struct {
	pool *pgxpool.Pool
}

var _ domain.TxManager = (*PgxTxManager)(nil)

// NewPgxTxManager construye el gestor transaccional con el pool de PostgreSQL.
func NewPgxTxManager(pool *pgxpool.Pool) *PgxTxManager {
	return &PgxTxManager{pool: pool}
}

// WithTransaction ejecuta la función provista dentro de una transacción atómica.
// Si fn retorna error o se genera un pánico en tiempo de ejecución, se ejecuta ROLLBACK automático.
// Si fn retorna nil, se ejecuta COMMIT confirmando las operaciones de manera indivisible.
func (m *PgxTxManager) WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) (err error) {
	if m.pool == nil {
		return fmt.Errorf("pool de base de datos no disponible para transacción")
	}

	tx, txErr := m.pool.Begin(ctx)
	if txErr != nil {
		return fmt.Errorf("error al abrir transacción con BEGIN: %w", txErr)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r) // Propagar el panic tras garantizar el rollback
		} else if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	txCtx := context.WithValue(ctx, txContextKey{}, tx)

	if err = fn(txCtx); err != nil {
		return err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("error al confirmar transacción con COMMIT: %w", commitErr)
	}

	return nil
}

// ExtractQuerier extrae la transacción activa asociada al contexto si existe; de lo contrario usa el querier por defecto.
func ExtractQuerier(ctx context.Context, defaultQ Querier) Querier {
	if tx, ok := ctx.Value(txContextKey{}).(pgx.Tx); ok && tx != nil {
		return tx
	}
	return defaultQ
}
