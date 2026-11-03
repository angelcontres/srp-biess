package database

import (
	"context"
	"encoding/json"
	"fmt"

	"srp-biess/internal/domain"
)

// PostgresNotificacionRepository implementa domain.NotificacionRepository con pgx y Querier (SRP-005, SDD 10.9).
type PostgresNotificacionRepository struct {
	q Querier
}

func NewPostgresNotificacionRepository(q Querier) *PostgresNotificacionRepository {
	return &PostgresNotificacionRepository{q: q}
}

func (r *PostgresNotificacionRepository) WithQuerier(q Querier) *PostgresNotificacionRepository {
	return &PostgresNotificacionRepository{q: q}
}

func (r *PostgresNotificacionRepository) getQuerier(ctx context.Context) Querier {
	return ExtractQuerier(ctx, r.q)
}

func (r *PostgresNotificacionRepository) Enqueue(ctx context.Context, n *domain.Notificacion) error {
	payloadJSON, err := json.Marshal(n.Payload)
	if err != nil {
		return fmt.Errorf("error serializando payload de notificación: %w", err)
	}

	query := `
		INSERT INTO notificacion (
			postulacion_id, postulante_id, tipo, destinatario, asunto, plantilla, payload, estado, fecha_creacion
		) VALUES (
			$1::uuid, $2::uuid, $3, $4, $5, $6, $7::jsonb, 'PENDIENTE', now()
		) RETURNING id, fecha_creacion`

	return r.getQuerier(ctx).QueryRow(ctx, query,
		n.PostulacionID,
		n.PostulanteID,
		n.Tipo,
		n.Destinatario,
		n.Asunto,
		n.Plantilla,
		string(payloadJSON),
	).Scan(&n.ID, &n.CreatedAt)
}
