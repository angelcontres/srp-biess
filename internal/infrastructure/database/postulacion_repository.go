package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"srp-biess/internal/domain"
)

// PostgresPostulacionRepository implementa domain.PostulacionRepository con pgx y Querier (SRP-004).
type PostgresPostulacionRepository struct {
	q Querier
}

func NewPostgresPostulacionRepository(q Querier) *PostgresPostulacionRepository {
	return &PostgresPostulacionRepository{q: q}
}

func (r *PostgresPostulacionRepository) WithQuerier(q Querier) *PostgresPostulacionRepository {
	return &PostgresPostulacionRepository{q: q}
}

func (r *PostgresPostulacionRepository) GetByID(ctx context.Context, id string) (*domain.Postulacion, error) {
	query := `
		SELECT id, COALESCE(codigo_unico, ''), postulante_id, COALESCE(cargo_id, 0), estado, created_at, updated_at
		FROM postulacion
		WHERE id = $1::uuid`

	var p domain.Postulacion
	var cargoID int
	var estado string
	var createdAt, updatedAt time.Time

	err := r.q.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.CodigoPostulacion,
		&p.PostulanteID,
		&cargoID,
		&estado,
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
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt

	return &p, nil
}

func (r *PostgresPostulacionRepository) GetByPostulanteID(ctx context.Context, postulanteID string) (*domain.Postulacion, error) {
	query := `
		SELECT id, COALESCE(codigo_unico, ''), postulante_id, COALESCE(cargo_id, 0), estado, created_at, updated_at
		FROM postulacion
		WHERE postulante_id = $1::uuid
		ORDER BY created_at DESC
		LIMIT 1`

	var p domain.Postulacion
	var cargoID int
	var estado string
	var createdAt, updatedAt time.Time

	err := r.q.QueryRow(ctx, query, postulanteID).Scan(
		&p.ID,
		&p.CodigoPostulacion,
		&p.PostulanteID,
		&cargoID,
		&estado,
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

	return r.q.QueryRow(ctx, query,
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

	cmdTag, err := r.q.Exec(ctx, query, id, string(estado))
	if err != nil {
		return fmt.Errorf("error actualizando estado de postulación: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
