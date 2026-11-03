package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"srp-biess/internal/domain"
)

// PostgresFormularioRepository también implementa domain.CargoRepository (SRP-017a):
// comparte la lectura del estado que consume la SPA.
var _ domain.CargoRepository = (*PostgresFormularioRepository)(nil)

func (r *PostgresFormularioRepository) ContextoConfirmacion(
	ctx context.Context, id, postulanteID string, cargoID int,
) (*domain.ContextoCargo, error) {
	if !idsValidos(id, postulanteID) {
		return nil, domain.ErrNotFound
	}
	if cargoID <= 0 || cargoID > 32767 { // cargo.id es smallint
		return nil, domain.ErrCargoInvalido
	}
	var (
		c      domain.ContextoCargo
		sector *string
	)
	err := r.q.QueryRow(ctx, `
		SELECT p.estado::text, p.version, p.elegibilidad_estado, pe.cedula, c.sector::text
		FROM postulacion p
		JOIN postulante pe ON pe.id = p.postulante_id
		LEFT JOIN cargo c ON c.id = $3::smallint AND c.activo
		WHERE p.id = $1::uuid AND p.postulante_id = $2::uuid`,
		id, postulanteID, cargoID,
	).Scan(&c.Estado, &c.Version, &c.ElegibilidadEstado, &c.Cedula, &sector)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, r.noEncontradaOAjena(ctx, id)
	}
	if err != nil {
		return nil, fmt.Errorf("error consultando la postulación para el cargo: %w", err)
	}
	if sector == nil {
		return nil, domain.ErrCargoInvalido
	}
	c.Sector = domain.SectorCargo(*sector)
	return &c, nil
}

// ConfirmarCargo escribe cargo, elegibilidad y, si no es elegible, el estado NO_ELEGIBLE en una
// sola sentencia. Solo aplica sobre un BORRADOR con elegibilidad PENDIENTE y la versión esperada.
func (r *PostgresFormularioRepository) ConfirmarCargo(
	ctx context.Context, id, postulanteID string, res domain.ResultadoElegibilidad, version int,
) error {
	if !idsValidos(id, postulanteID) {
		return domain.ErrNotFound
	}
	elegibilidad := domain.ElegibilidadNoElegible
	if res.Elegible {
		elegibilidad = domain.ElegibilidadElegible
	}
	q := ExtractQuerier(ctx, r.q)
	var actualizada string
	err := q.QueryRow(ctx, `
		UPDATE postulacion
		SET cargo_id = $3::smallint,
		    elegibilidad_estado = $4::varchar,
		    elegibilidad_motivo = $5,
		    estado = CASE WHEN $4::varchar = 'NO_ELEGIBLE' THEN 'NO_ELEGIBLE'::estado_postulacion ELSE estado END,
		    version = version + 1,
		    updated_at = now()
		WHERE id = $1::uuid AND postulante_id = $2::uuid AND version = $6
		  AND estado = 'BORRADOR' AND elegibilidad_estado = 'PENDIENTE'
		RETURNING id`,
		id, postulanteID, res.CargoID, elegibilidad, res.Motivo, version,
	).Scan(&actualizada)
	if errors.Is(err, pgx.ErrNoRows) {
		return r.diagnosticarRechazoCargo(ctx, q, id, postulanteID)
	}
	if err != nil {
		return fmt.Errorf("error guardando el cargo: %w", err)
	}
	return nil
}

func (r *PostgresFormularioRepository) EstadoPostulacion(ctx context.Context, id, postulanteID string) (*domain.EstadoFormulario, error) {
	if !idsValidos(id, postulanteID) {
		return nil, domain.ErrNotFound
	}
	return r.leerEstado(ctx, "WHERE p.id = $1::uuid AND p.postulante_id = $2::uuid", id, postulanteID)
}

// noEncontradaOAjena distingue una postulación inexistente (404) de una ajena (403).
func (r *PostgresFormularioRepository) noEncontradaOAjena(ctx context.Context, id string) error {
	var existe bool
	if err := r.q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM postulacion WHERE id = $1::uuid)`, id).Scan(&existe); err != nil {
		return fmt.Errorf("error verificando postulación: %w", err)
	}
	if existe {
		return domain.ErrAccesoDenegado
	}
	return domain.ErrNotFound
}

// Otra sesión pudo confirmar el cargo o guardar una sección mientras se consultaba al IESS.
func (r *PostgresFormularioRepository) diagnosticarRechazoCargo(ctx context.Context, q Querier, id, postulanteID string) error {
	var estado, elegibilidad string
	err := q.QueryRow(ctx,
		`SELECT estado::text, elegibilidad_estado FROM postulacion WHERE id = $1::uuid AND postulante_id = $2::uuid`,
		id, postulanteID).Scan(&estado, &elegibilidad)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return r.noEncontradaOAjena(ctx, id)
	case err != nil:
		return fmt.Errorf("error verificando postulación: %w", err)
	case estado != "BORRADOR":
		return domain.ErrPostulacionCerrada
	case elegibilidad != domain.ElegibilidadPendiente:
		return domain.ErrCargoYaConfirmado
	default:
		return domain.ErrConflictoVersion
	}
}
