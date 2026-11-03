package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"srp-biess/internal/domain"
)

// PostgresPostulacionRepository implementa domain.PostulacionRepository con pgx y Querier (SRP-004, SRP-005).
type PostgresPostulacionRepository struct {
	q Querier
}

func NewPostgresPostulacionRepository(q Querier) *PostgresPostulacionRepository {
	return &PostgresPostulacionRepository{q: q}
}

func (r *PostgresPostulacionRepository) WithQuerier(q Querier) *PostgresPostulacionRepository {
	return &PostgresPostulacionRepository{q: q}
}

func (r *PostgresPostulacionRepository) getQuerier(ctx context.Context) Querier {
	return ExtractQuerier(ctx, r.q)
}

func (r *PostgresPostulacionRepository) GetByID(ctx context.Context, id string) (*domain.Postulacion, error) {
	return r.getByID(ctx, id, false)
}

func (r *PostgresPostulacionRepository) GetByIDForUpdate(ctx context.Context, id string) (*domain.Postulacion, error) {
	return r.getByID(ctx, id, true)
}

func (r *PostgresPostulacionRepository) getByID(ctx context.Context, id string, forUpdate bool) (*domain.Postulacion, error) {
	query := `
		SELECT id, COALESCE(codigo_unico, ''), postulante_id, COALESCE(cargo_id, 0), estado, fecha_envio, created_at, updated_at
		FROM postulacion
		WHERE id = $1::uuid`
	if forUpdate {
		query += " FOR UPDATE"
	}

	var p domain.Postulacion
	var cargoID int
	var estado string
	var fechaEnvio *time.Time
	var createdAt, updatedAt time.Time

	err := r.getQuerier(ctx).QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.CodigoPostulacion,
		&p.PostulanteID,
		&cargoID,
		&estado,
		&fechaEnvio,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error consultando postulacion por ID: %w", err)
	}

	p.Estado = domain.EstadoPostulacion(estado)
	if cargoID > 0 {
		p.CargoID = strconv.Itoa(cargoID)
	}
	p.FechaEnvio = fechaEnvio
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt

	return &p, nil
}

func (r *PostgresPostulacionRepository) GetByPostulanteID(ctx context.Context, postulanteID string) (*domain.Postulacion, error) {
	query := `
		SELECT id, COALESCE(codigo_unico, ''), postulante_id, COALESCE(cargo_id, 0), estado, fecha_envio, created_at, updated_at
		FROM postulacion
		WHERE postulante_id = $1::uuid
		ORDER BY created_at DESC
		LIMIT 1`

	var p domain.Postulacion
	var cargoID int
	var estado string
	var fechaEnvio *time.Time
	var createdAt, updatedAt time.Time

	err := r.getQuerier(ctx).QueryRow(ctx, query, postulanteID).Scan(
		&p.ID,
		&p.CodigoPostulacion,
		&p.PostulanteID,
		&cargoID,
		&estado,
		&fechaEnvio,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error consultando postulacion por postulante_id: %w", err)
	}

	p.Estado = domain.EstadoPostulacion(estado)
	if cargoID > 0 {
		p.CargoID = strconv.Itoa(cargoID)
	}
	p.FechaEnvio = fechaEnvio
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt

	return &p, nil
}

func (r *PostgresPostulacionRepository) Save(ctx context.Context, p *domain.Postulacion) error {
	var cargoID sql.NullInt32
	if p.CargoID != "" {
		if c, err := strconv.Atoi(p.CargoID); err == nil && c > 0 {
			cargoID = sql.NullInt32{Int32: int32(c), Valid: true}
		}
	}

	var codigo sql.NullString
	if p.CodigoPostulacion != "" {
		codigo = sql.NullString{String: p.CodigoPostulacion, Valid: true}
	}

	query := `
		INSERT INTO postulacion (
			codigo_unico, postulante_id, convocatoria_id, cargo_id, estado, created_at, updated_at
		) VALUES (
			$1, $2::uuid, 1, $3, $4::estado_postulacion, now(), now()
		) RETURNING id, created_at, updated_at`

	estado := string(p.Estado)
	if estado == "" {
		estado = string(domain.EstadoBorrador)
	}

	return r.getQuerier(ctx).QueryRow(ctx, query,
		codigo,
		p.PostulanteID,
		cargoID,
		estado,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PostgresPostulacionRepository) UpdateEstado(ctx context.Context, id string, estado domain.EstadoPostulacion) error {
	query := `
		UPDATE postulacion
		SET estado = $2::estado_postulacion, updated_at = now()
		WHERE id = $1::uuid`

	cmdTag, err := r.getQuerier(ctx).Exec(ctx, query, id, string(estado))
	if err != nil {
		return fmt.Errorf("error actualizando estado de postulación: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateSubmit actualiza el estado a ENVIADA, asigna el código único y la fecha de envío en una sola sentencia atómica (SRP-005, SDD 10.9).
func (r *PostgresPostulacionRepository) UpdateSubmit(ctx context.Context, id, codigoUnico string, fechaEnvio time.Time) error {
	query := `
		UPDATE postulacion
		SET estado = 'ENVIADA',
		    codigo_unico = $2,
		    fecha_envio = $3,
		    updated_at = now()
		WHERE id = $1::uuid AND estado IN ('BORRADOR', 'LISTA_PARA_ENVIO')`

	cmdTag, err := r.getQuerier(ctx).Exec(ctx, query, id, codigoUnico, fechaEnvio)
	if err != nil {
		return fmt.Errorf("error actualizando envio de postulación: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrPostulacionYaEnviada
	}
	return nil
}

// NextCodigoUnico genera el siguiente código correlativo único institucional usando la secuencia de PostgreSQL (SRP-005, SDD 10.9).
func (r *PostgresPostulacionRepository) NextCodigoUnico(ctx context.Context) (string, error) {
	query := `SELECT nextval('seq_codigo_postulacion')`
	var seqVal int64
	err := r.getQuerier(ctx).QueryRow(ctx, query).Scan(&seqVal)
	if err != nil {
		return "", fmt.Errorf("error generando correlativo de secuencia seq_codigo_postulacion: %w", err)
	}
	year := time.Now().Year()
	return fmt.Sprintf("BIESS-%d-%05d", year, seqVal), nil
}

func (r *PostgresPostulacionRepository) GetIdempotentSubmission(ctx context.Context, key string) (*domain.IdempotentSubmission, error) {
	var record domain.IdempotentSubmission
	var response []byte
	err := r.getQuerier(ctx).QueryRow(ctx, `
		SELECT idempotency_key, postulacion_id::text, response
		FROM postulacion_envio_idempotencia
		WHERE idempotency_key = $1
		FOR UPDATE`, key).Scan(&record.Key, &record.PostulacionID, &response)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error consultando idempotencia: %w", err)
	}
	record.Response = response
	return &record, nil
}

func (r *PostgresPostulacionRepository) SaveIdempotentSubmission(ctx context.Context, key, postulacionID string, response []byte) error {
	_, err := r.getQuerier(ctx).Exec(ctx, `
		INSERT INTO postulacion_envio_idempotencia (idempotency_key, postulacion_id, response)
		VALUES ($1, $2::uuid, $3::jsonb)`, key, postulacionID, string(response))
	if err != nil {
		return fmt.Errorf("error guardando idempotencia de envío: %w", err)
	}
	return nil
}

func (r *PostgresPostulacionRepository) SaveConsentimiento(ctx context.Context, postulacionID, version string) error {
	_, err := r.getQuerier(ctx).Exec(ctx, `
		INSERT INTO consentimiento (postulacion_id, version_terminos, aceptado, textos_aceptados)
		VALUES ($1::uuid, $2, true, '{}'::jsonb)
		ON CONFLICT (postulacion_id) DO UPDATE
		SET version_terminos = EXCLUDED.version_terminos,
		    aceptado = EXCLUDED.aceptado,
		    fecha_aceptacion = now()`, postulacionID, version)
	if err != nil {
		return fmt.Errorf("error guardando consentimiento: %w", err)
	}
	return nil
}

// SaveTextosConsentimiento guarda los textos aceptados (SDD 8.5.8); el PDF imprime estos mismos.
func (r *PostgresPostulacionRepository) SaveTextosConsentimiento(ctx context.Context, postulacionID string, textos []string) error {
	datos, err := json.Marshal(textos)
	if err != nil {
		return fmt.Errorf("error serializando consentimientos: %w", err)
	}
	tag, err := r.getQuerier(ctx).Exec(ctx,
		`UPDATE consentimiento SET textos_aceptados = $2::jsonb WHERE postulacion_id = $1::uuid`, postulacionID, string(datos))
	if err != nil {
		return fmt.Errorf("error guardando los textos de consentimiento: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
