package database

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"srp-biess/internal/domain"
)

// PostgresFormularioRepository implementa domain.FormularioRepository (SRP-013, integración mínima).
// Las secciones se guardan como JSONB en `postulacion_seccion`
// (migración db/migrations/001_postulacion_seccion.sql).
type PostgresFormularioRepository struct {
	q Querier
}

func NewPostgresFormularioRepository(q Querier) *PostgresFormularioRepository {
	return &PostgresFormularioRepository{q: q}
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Evita errores de cast en PostgreSQL con identificadores mal formados: se tratan como no encontrados.
func idsValidos(ids ...string) bool {
	for _, id := range ids {
		if !uuidPattern.MatchString(id) {
			return false
		}
	}
	return true
}

const selectEstado = `
	SELECT p.id, p.estado::text, p.version, p.fecha_guardado, COALESCE(p.ultima_seccion, ''),
	       COALESCE(p.cargo_id, 0), p.elegibilidad_estado, COALESCE(p.elegibilidad_motivo, ''), p.updated_at
	FROM postulacion p`

func (r *PostgresFormularioRepository) leerEstado(ctx context.Context, where string, args ...any) (*domain.EstadoFormulario, error) {
	var (
		e                  domain.EstadoFormulario
		fechaGuardado      *time.Time
		ultimaSeccion      string
		cargoID            int
		elegibilidadEstado string
		motivo             string
		actualizado        time.Time
	)
	err := r.q.QueryRow(ctx, selectEstado+" "+where, args...).Scan(
		&e.ID, &e.Estado, &e.Version, &fechaGuardado, &ultimaSeccion,
		&cargoID, &elegibilidadEstado, &motivo, &actualizado,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error consultando postulación: %w", err)
	}

	e.UltimoGuardado = fechaGuardado
	e.SeccionReanudacion = "personal"
	if domain.EsSeccionFormulario(ultimaSeccion) {
		e.SeccionReanudacion = ultimaSeccion
	}
	if cargoID > 0 && (elegibilidadEstado == "ELEGIBLE" || elegibilidadEstado == "NO_ELEGIBLE") {
		e.Elegibilidad = &domain.Elegibilidad{
			Elegible: elegibilidadEstado == "ELEGIBLE",
			CargoID:  cargoID,
			Mensaje:  motivo,
			Fecha:    actualizado,
		}
	}

	e.Secciones = map[string]domain.EstadoSeccion{"resumen": domain.SeccionSinIniciar}
	for _, s := range domain.SeccionesFormulario {
		e.Secciones[s] = domain.SeccionSinIniciar
	}
	rows, err := r.q.Query(ctx, `SELECT seccion, completa FROM postulacion_seccion WHERE postulacion_id = $1::uuid`, e.ID)
	if err != nil {
		return nil, fmt.Errorf("error consultando secciones: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var seccion string
		var completa bool
		if err := rows.Scan(&seccion, &completa); err != nil {
			return nil, fmt.Errorf("error leyendo secciones: %w", err)
		}
		if completa {
			e.Secciones[seccion] = domain.SeccionCompleta
		} else {
			e.Secciones[seccion] = domain.SeccionIncompleta
		}
	}
	return &e, rows.Err()
}

func (r *PostgresFormularioRepository) ObtenerActual(ctx context.Context, postulanteID string) (*domain.EstadoFormulario, error) {
	if !idsValidos(postulanteID) {
		return nil, domain.ErrNotFound
	}
	return r.leerEstado(ctx, "WHERE p.postulante_id = $1::uuid ORDER BY p.created_at DESC LIMIT 1", postulanteID)
}

func (r *PostgresFormularioRepository) CrearBorrador(ctx context.Context, postulanteID string) (*domain.EstadoFormulario, error) {
	if !idsValidos(postulanteID) {
		return nil, domain.ErrNotFound
	}
	// Solo en una convocatoria con la fase habilitada y dentro de sus fechas (RF-EF12-01).
	var id string
	err := r.q.QueryRow(ctx, `
		INSERT INTO postulacion (postulante_id, convocatoria_id)
		SELECT $1::uuid, c.id
		FROM convocatoria c
		WHERE c.fase_postulacion_habilitada
		  AND current_date BETWEEN c.fecha_inicio_postulacion AND c.fecha_fin_postulacion
		ORDER BY c.id DESC
		LIMIT 1
		RETURNING id`, postulanteID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFaseCerrada
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrPostulacionExists
		}
		return nil, fmt.Errorf("error creando borrador: %w", err)
	}
	return r.leerEstado(ctx, "WHERE p.id = $1::uuid", id)
}

func (r *PostgresFormularioRepository) ObtenerDetalle(ctx context.Context, id, postulanteID string) (*domain.DetalleFormulario, error) {
	if !idsValidos(id, postulanteID) {
		return nil, domain.ErrNotFound
	}
	estado, err := r.leerEstado(ctx, "WHERE p.id = $1::uuid AND p.postulante_id = $2::uuid", id, postulanteID)
	if err != nil {
		return nil, err
	}

	var v domain.DatosVerificados
	err = r.q.QueryRow(ctx, `
		SELECT cedula, nombres, apellidos, COALESCE(genero::text, ''),
		       COALESCE(to_char(fecha_nacimiento, 'YYYY-MM-DD'), ''), nacionalidad, correo
		FROM postulante WHERE id = $1::uuid`, postulanteID).
		Scan(&v.Cedula, &v.Nombres, &v.Apellidos, &v.Genero, &v.FechaNacimiento, &v.Nacionalidad, &v.Correo)
	if err != nil {
		return nil, fmt.Errorf("error consultando datos verificados: %w", err)
	}

	secciones := map[string][]byte{}
	rows, err := r.q.Query(ctx, `SELECT seccion, datos::text FROM postulacion_seccion WHERE postulacion_id = $1::uuid`, id)
	if err != nil {
		return nil, fmt.Errorf("error consultando valores de secciones: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var seccion, datos string
		if err := rows.Scan(&seccion, &datos); err != nil {
			return nil, fmt.Errorf("error leyendo valores de secciones: %w", err)
		}
		secciones[seccion] = []byte(datos)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.DetalleFormulario{Postulacion: estado, Verificados: &v, Secciones: secciones}, nil
}

// GuardarSeccion actualiza versión y sección en una sola sentencia (atómica sin TxManager):
// si la versión no coincide o la postulación no está en BORRADOR, no se escribe nada.
func (r *PostgresFormularioRepository) GuardarSeccion(
	ctx context.Context, id, postulanteID, seccion string, datos []byte, completa bool, version int,
) (*domain.EstadoFormulario, error) {
	if !idsValidos(id, postulanteID) {
		return nil, domain.ErrNotFound
	}
	var guardada string
	err := r.q.QueryRow(ctx, `
		WITH actualizada AS (
			UPDATE postulacion
			SET version = version + 1, ultima_seccion = $3, fecha_guardado = now(), updated_at = now()
			WHERE id = $1::uuid AND postulante_id = $2::uuid AND version = $6 AND estado = 'BORRADOR'
			RETURNING id
		)
		INSERT INTO postulacion_seccion (postulacion_id, seccion, datos, completa, updated_at)
		SELECT id, $3, $4::jsonb, $5, now() FROM actualizada
		ON CONFLICT (postulacion_id, seccion)
		DO UPDATE SET datos = EXCLUDED.datos, completa = EXCLUDED.completa, updated_at = now()
		RETURNING postulacion_id`,
		id, postulanteID, seccion, string(datos), completa, version,
	).Scan(&guardada)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, r.diagnosticarRechazo(ctx, id, postulanteID)
	}
	if err != nil {
		return nil, fmt.Errorf("error guardando sección %s: %w", seccion, err)
	}
	return r.leerEstado(ctx, "WHERE p.id = $1::uuid", id)
}

func (r *PostgresFormularioRepository) diagnosticarRechazo(ctx context.Context, id, postulanteID string) error {
	var estado string
	err := r.q.QueryRow(ctx,
		`SELECT estado::text FROM postulacion WHERE id = $1::uuid AND postulante_id = $2::uuid`,
		id, postulanteID).Scan(&estado)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.ErrNotFound
	case err != nil:
		return fmt.Errorf("error verificando postulación: %w", err)
	case estado != "BORRADOR":
		return domain.ErrPostulacionCerrada
	default:
		return domain.ErrConflictoVersion
	}
}
