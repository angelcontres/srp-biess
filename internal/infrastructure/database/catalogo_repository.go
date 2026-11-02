package database

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"

	"srp-biess/internal/domain"
)

// PostgresCatalogoRepository implementa domain.CatalogoRepository usando pgx/v5 y Querier (SRP-004).
type PostgresCatalogoRepository struct {
	q Querier
}

// NewPostgresCatalogoRepository instancia el repositorio inyectando un Querier (Pool o Tx).
func NewPostgresCatalogoRepository(q Querier) *PostgresCatalogoRepository {
	return &PostgresCatalogoRepository{q: q}
}

// WithQuerier permite ejecutar operaciones sobre una transacción específica.
func (r *PostgresCatalogoRepository) WithQuerier(q Querier) *PostgresCatalogoRepository {
	return &PostgresCatalogoRepository{q: q}
}

func (r *PostgresCatalogoRepository) GetPaises(ctx context.Context) ([]domain.Pais, error) {
	query := `SELECT id, codigo_iso2, nombre FROM pais WHERE activo = true ORDER BY nombre ASC`
	rows, err := r.q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error consultando paises: %w", err)
	}
	defer rows.Close()

	var paises []domain.Pais
	for rows.Next() {
		var id int
		var iso2, nombre string
		if err := rows.Scan(&id, &iso2, &nombre); err != nil {
			return nil, fmt.Errorf("error escaneando pais: %w", err)
		}
		paises = append(paises, domain.Pais{
			ID:     strconv.Itoa(id),
			Codigo: iso2,
			Nombre: nombre,
		})
	}

	return paises, rows.Err()
}

func (r *PostgresCatalogoRepository) GetProvincias(ctx context.Context, paisID string) ([]domain.Provincia, error) {
	query := `SELECT id, pais_id, codigo, nombre FROM provincia WHERE pais_id = $1 AND activo = true ORDER BY nombre ASC`
	pid, err := strconv.Atoi(paisID)
	if err != nil {
		return nil, fmt.Errorf("pais_id inválido: %w", err)
	}

	rows, err := r.q.Query(ctx, query, pid)
	if err != nil {
		return nil, fmt.Errorf("error consultando provincias: %w", err)
	}
	defer rows.Close()

	var provincias []domain.Provincia
	for rows.Next() {
		var id, pID int
		var codigo, nombre string
		if err := rows.Scan(&id, &pID, &codigo, &nombre); err != nil {
			return nil, fmt.Errorf("error escaneando provincia: %w", err)
		}
		provincias = append(provincias, domain.Provincia{
			ID:     strconv.Itoa(id),
			PaisID: strconv.Itoa(pID),
			Codigo: codigo,
			Nombre: nombre,
		})
	}

	return provincias, rows.Err()
}

func (r *PostgresCatalogoRepository) GetCantones(ctx context.Context, provinciaID string) ([]domain.Canton, error) {
	query := `SELECT id, provincia_id, nombre FROM canton WHERE provincia_id = $1 AND activo = true ORDER BY nombre ASC`
	prID, err := strconv.Atoi(provinciaID)
	if err != nil {
		return nil, fmt.Errorf("provincia_id inválido: %w", err)
	}

	rows, err := r.q.Query(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("error consultando cantones: %w", err)
	}
	defer rows.Close()

	var cantones []domain.Canton
	for rows.Next() {
		var id, provID int
		var nombre string
		if err := rows.Scan(&id, &provID, &nombre); err != nil {
			return nil, fmt.Errorf("error escaneando canton: %w", err)
		}
		cantones = append(cantones, domain.Canton{
			ID:          strconv.Itoa(id),
			ProvinciaID: strconv.Itoa(provID),
			Nombre:      nombre,
		})
	}

	return cantones, rows.Err()
}

func (r *PostgresCatalogoRepository) GetCargos(ctx context.Context) ([]domain.Cargo, error) {
	query := `SELECT id, codigo, nombre, activo FROM cargo WHERE activo = true ORDER BY id ASC`
	rows, err := r.q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error consultando cargos: %w", err)
	}
	defer rows.Close()

	var cargos []domain.Cargo
	for rows.Next() {
		var id int
		var codigo, nombre string
		var activo bool
		if err := rows.Scan(&id, &codigo, &nombre, &activo); err != nil {
			return nil, fmt.Errorf("error escaneando cargo: %w", err)
		}
		cargos = append(cargos, domain.Cargo{
			ID:          strconv.Itoa(id),
			Nombre:      nombre,
			Descripcion: codigo,
			Activo:      activo,
		})
	}

	return cargos, rows.Err()
}

func (r *PostgresCatalogoRepository) GetCargoByID(ctx context.Context, id string) (*domain.Cargo, error) {
	query := `SELECT id, codigo, nombre, activo FROM cargo WHERE id = $1`
	cID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("id de cargo inválido: %w", err)
	}

	var cargoID int
	var codigo, nombre string
	var activo bool
	err = r.q.QueryRow(ctx, query, cID).Scan(&cargoID, &codigo, &nombre, &activo)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error consultando cargo por ID: %w", err)
	}

	return &domain.Cargo{
		ID:          strconv.Itoa(cargoID),
		Nombre:      nombre,
		Descripcion: codigo,
		Activo:      activo,
	}, nil
}
