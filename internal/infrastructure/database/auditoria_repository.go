package database

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"srp-biess/internal/domain"
)

// PostgresAuditoriaRepository implementa domain.AuditoriaRepository sobre `auditoria_evento`.
type PostgresAuditoriaRepository struct {
	q Querier
}

var _ domain.AuditoriaRepository = (*PostgresAuditoriaRepository)(nil)

func NewPostgresAuditoriaRepository(q Querier) *PostgresAuditoriaRepository {
	return &PostgresAuditoriaRepository{q: q}
}

func (r *PostgresAuditoriaRepository) Registrar(ctx context.Context, e domain.EventoAuditoria) error {
	antes, err := jsonONulo(e.Antes)
	if err != nil {
		return err
	}
	despues, err := jsonONulo(e.Despues)
	if err != nil {
		return err
	}
	_, err = ExtractQuerier(ctx, r.q).Exec(ctx, `
		INSERT INTO auditoria_evento
			(actor_id, actor_tipo, accion, entidad, entidad_id, datos_antes, datos_despues, ip, user_agent, trace_id)
		VALUES ($1::uuid, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8::inet, $9, $10::uuid)`,
		uuidONulo(e.ActorID), e.ActorTipo, e.Accion, e.Entidad, e.EntidadID,
		antes, despues, ipONula(e.Solicitud.IP), textoONulo(e.Solicitud.UserAgent), uuidONulo(e.Solicitud.TraceID),
	)
	if err != nil {
		return fmt.Errorf("error registrando auditoría %s: %w", e.Accion, err)
	}
	return nil
}

func jsonONulo(v map[string]any) (*string, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("error serializando auditoría: %w", err)
	}
	s := string(b)
	return &s, nil
}

// Un valor mal formado se guarda como NULL: la auditoría no debe fallar por un dato de origen.
func uuidONulo(s string) *string {
	if !uuidPattern.MatchString(s) {
		return nil
	}
	return &s
}

func ipONula(s string) *string {
	if net.ParseIP(s) == nil {
		return nil
	}
	return &s
}

func textoONulo(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
