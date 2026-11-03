package domain

import "context"

// TxManager define el contrato para la orquestación de transacciones atómicas (ADR-004, SRP-005).
type TxManager interface {
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

type NoopTxManager struct{}

func (NoopTxManager) WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}
