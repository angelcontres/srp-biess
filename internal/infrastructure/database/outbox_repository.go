package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"srp-biess/internal/domain"
)

// PostgresNotificacionRepository también implementa domain.OutboxRepository (ADR-011).
var _ domain.OutboxRepository = (*PostgresNotificacionRepository)(nil)

// TomarPendiente bloquea la notificación más antigua lista para enviarse. SKIP LOCKED: si otro
// worker ya tiene una, se salta y toma la siguiente, así ninguna se envía dos veces a la vez.
func (r *PostgresNotificacionRepository) TomarPendiente(ctx context.Context) (*domain.NotificacionEnCola, error) {
	var (
		n        domain.NotificacionEnCola
		payload  []byte
		postul   *string
		postulte *string
	)
	err := r.getQuerier(ctx).QueryRow(ctx, `
		SELECT id, postulacion_id::text, postulante_id::text, tipo, destinatario, cc, asunto, plantilla,
		       payload, estado::text, intentos, fecha_creacion
		FROM notificacion
		WHERE estado = 'PENDIENTE' AND proximo_intento <= now()
		ORDER BY fecha_creacion
		LIMIT 1
		FOR UPDATE SKIP LOCKED`,
	).Scan(&n.ID, &postul, &postulte, &n.Tipo, &n.Destinatario, &n.CC, &n.Asunto, &n.Plantilla,
		&payload, &n.Estado, &n.Intentos, &n.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error tomando notificación pendiente: %w", err)
	}
	if postul != nil {
		n.PostulacionID = *postul
	}
	if postulte != nil {
		n.PostulanteID = *postulte
	}
	if err := json.Unmarshal(payload, &n.Payload); err != nil {
		n.Payload = map[string]any{}
	}
	return &n, nil
}

func (r *PostgresNotificacionRepository) MarcarEnviada(ctx context.Context, id string) error {
	_, err := r.getQuerier(ctx).Exec(ctx, `
		UPDATE notificacion SET estado = 'ENVIADA', fecha_envio = now(), intentos = intentos + 1, error = NULL
		WHERE id = $1::uuid`, id)
	if err != nil {
		return fmt.Errorf("error marcando notificación enviada: %w", err)
	}
	return nil
}

func (r *PostgresNotificacionRepository) ReprogramarFallo(
	ctx context.Context, id string, intentos int, proximo time.Time, motivo string, definitivo bool,
) error {
	estado := "PENDIENTE"
	if definitivo {
		estado = "FALLIDA"
	}
	if r := []rune(motivo); len(r) > 1000 {
		motivo = string(r[:1000])
	}
	_, err := r.getQuerier(ctx).Exec(ctx, `
		UPDATE notificacion
		SET intentos = $2, proximo_intento = $3, error = $4, estado = $5::estado_notificacion
		WHERE id = $1::uuid`, id, intentos, proximo, motivo, estado)
	if err != nil {
		return fmt.Errorf("error reprogramando notificación: %w", err)
	}
	return nil
}
