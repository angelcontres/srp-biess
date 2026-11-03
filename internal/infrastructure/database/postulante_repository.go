package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"srp-biess/internal/domain"
)

// PostgresPostulanteRepository implementa domain.PostulanteRepository con pgx y Querier (SRP-004, SRP-005).
type PostgresPostulanteRepository struct {
	q Querier
}

func NewPostgresPostulanteRepository(q Querier) *PostgresPostulanteRepository {
	return &PostgresPostulanteRepository{q: q}
}

func (r *PostgresPostulanteRepository) WithQuerier(q Querier) *PostgresPostulanteRepository {
	return &PostgresPostulanteRepository{q: q}
}

func (r *PostgresPostulanteRepository) getQuerier(ctx context.Context) Querier {
	return ExtractQuerier(ctx, r.q)
}

func (r *PostgresPostulanteRepository) GetByID(ctx context.Context, id string) (*domain.Postulante, error) {
	query := `
		SELECT id, cedula, nombres, apellidos, correo, password_hash, created_at, updated_at
		FROM postulante
		WHERE id = $1::uuid`

	var p domain.Postulante
	var createdAt, updatedAt time.Time

	err := r.getQuerier(ctx).QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Cedula,
		&p.Nombres,
		&p.Apellidos,
		&p.Email,
		&p.PasswordHash,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error consultando postulante por ID: %w", err)
	}

	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt
	return &p, nil
}

func (r *PostgresPostulanteRepository) GetByCedula(ctx context.Context, cedula string) (*domain.Postulante, error) {
	query := `
		SELECT id, cedula, nombres, apellidos, correo, password_hash, created_at, updated_at
		FROM postulante
		WHERE cedula = $1`

	var p domain.Postulante
	var createdAt, updatedAt time.Time

	err := r.getQuerier(ctx).QueryRow(ctx, query, cedula).Scan(
		&p.ID,
		&p.Cedula,
		&p.Nombres,
		&p.Apellidos,
		&p.Email,
		&p.PasswordHash,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error consultando postulante por cedula: %w", err)
	}

	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt
	return &p, nil
}

func (r *PostgresPostulanteRepository) GetByEmail(ctx context.Context, email string) (*domain.Postulante, error) {
	query := `
		SELECT id, cedula, nombres, apellidos, correo, password_hash, created_at, updated_at
		FROM postulante
		WHERE correo = $1`

	var p domain.Postulante
	var createdAt, updatedAt time.Time

	err := r.getQuerier(ctx).QueryRow(ctx, query, email).Scan(
		&p.ID,
		&p.Cedula,
		&p.Nombres,
		&p.Apellidos,
		&p.Email,
		&p.PasswordHash,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error consultando postulante por email: %w", err)
	}

	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt
	return &p, nil
}

// Save guarda también los datos del Registro Civil (género, fecha de nacimiento, nacionalidad y
// estado civil). Sin ellos, los datos verificados y el formulario PDF quedan incompletos.
func (r *PostgresPostulanteRepository) Save(ctx context.Context, p *domain.Postulante) error {
	query := `
		INSERT INTO postulante (
			cedula, nombres, apellidos, nacionalidad, correo, password_hash, estado,
			genero, fecha_nacimiento, estado_civil, created_at, updated_at
		) VALUES (
			$1, $2, $3, COALESCE(NULLIF($6, ''), 'ECUATORIANA'), $4, $5, 'ACTIVO',
			NULLIF($7, '')::genero, $8::date, NULLIF($9, '')::estado_civil, now(), now()
		) RETURNING id, created_at, updated_at`

	return r.getQuerier(ctx).QueryRow(ctx, query,
		p.Cedula,
		p.Nombres,
		p.Apellidos,
		p.Email,
		p.PasswordHash,
		p.Nacionalidad,
		p.Genero,
		p.FechaNacimiento,
		p.EstadoCivil,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PostgresPostulanteRepository) Update(ctx context.Context, p *domain.Postulante) error {
	query := `
		UPDATE postulante
		SET nombres = $2, apellidos = $3, correo = $4, updated_at = now()
		WHERE id = $1::uuid`

	cmdTag, err := r.getQuerier(ctx).Exec(ctx, query, p.ID, p.Nombres, p.Apellidos, p.Email)
	if err != nil {
		return fmt.Errorf("error actualizando postulante: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
