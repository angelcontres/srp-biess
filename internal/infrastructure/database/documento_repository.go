package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"srp-biess/internal/domain"
)

// PostgresDocumentoRepository implementa domain.DocumentoRepository con pgx y Querier (SRP-004).
type PostgresDocumentoRepository struct {
	q Querier
}

func NewPostgresDocumentoRepository(q Querier) *PostgresDocumentoRepository {
	return &PostgresDocumentoRepository{q: q}
}

func (r *PostgresDocumentoRepository) WithQuerier(q Querier) *PostgresDocumentoRepository {
	return &PostgresDocumentoRepository{q: q}
}

func (r *PostgresDocumentoRepository) Save(ctx context.Context, doc *domain.Documento) error {
	var postulacionID sql.NullString
	if doc.PostulacionID != "" {
		postulacionID = sql.NullString{String: doc.PostulacionID, Valid: true}
	}

	query := `
		INSERT INTO documento (
			postulante_id, postulacion_id, tipo, nombre_original, nombre_almacenado,
			mime_type, tamano_bytes, hash_sha256, ruta, estado, created_at
		) VALUES (
			$1::uuid, $2, $3::tipo_documento, $4, $5,
			$6, $7, $8, $9, $10::estado_documento, now()
		) RETURNING id, created_at`

	estado := string(doc.Estado)
	if estado == "" {
		estado = string(domain.EstadoDocPendiente)
	}

	return r.q.QueryRow(ctx, query,
		doc.PostulanteID,
		postulacionID,
		string(doc.Tipo),
		doc.NombreOriginal,
		doc.NombreAlmacenado,
		doc.MimeType,
		doc.TamanoBytes,
		doc.HashSHA256,
		doc.Ruta,
		estado,
	).Scan(&doc.ID, &doc.CreatedAt)
}

func (r *PostgresDocumentoRepository) GetByID(ctx context.Context, id string) (*domain.Documento, error) {
	query := `
		SELECT id, postulante_id, COALESCE(postulacion_id::text, ''), tipo, nombre_original,
		       nombre_almacenado, mime_type, tamano_bytes, hash_sha256, ruta, estado, created_at
		FROM documento
		WHERE id = $1::uuid`

	var d domain.Documento
	var tipo, estado string
	var createdAt time.Time

	err := r.q.QueryRow(ctx, query, id).Scan(
		&d.ID,
		&d.PostulanteID,
		&d.PostulacionID,
		&tipo,
		&d.NombreOriginal,
		&d.NombreAlmacenado,
		&d.MimeType,
		&d.TamanoBytes,
		&d.HashSHA256,
		&d.Ruta,
		&estado,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error consultando documento por ID: %w", err)
	}

	d.Tipo = domain.TipoDocumento(tipo)
	d.Estado = domain.EstadoDocumento(estado)
	d.CreatedAt = createdAt

	return &d, nil
}

func (r *PostgresDocumentoRepository) ListByPostulacionID(ctx context.Context, postulacionID string) ([]domain.Documento, error) {
	query := `
		SELECT id, postulante_id, COALESCE(postulacion_id::text, ''), tipo, nombre_original,
		       nombre_almacenado, mime_type, tamano_bytes, hash_sha256, ruta, estado, created_at
		FROM documento
		WHERE postulacion_id = $1::uuid
		ORDER BY created_at ASC`

	rows, err := r.q.Query(ctx, query, postulacionID)
	if err != nil {
		return nil, fmt.Errorf("error listando documentos de postulación: %w", err)
	}
	defer rows.Close()

	var docs []domain.Documento
	for rows.Next() {
		var d domain.Documento
		var tipo, estado string
		var createdAt time.Time

		err := rows.Scan(
			&d.ID,
			&d.PostulanteID,
			&d.PostulacionID,
			&tipo,
			&d.NombreOriginal,
			&d.NombreAlmacenado,
			&d.MimeType,
			&d.TamanoBytes,
			&d.HashSHA256,
			&d.Ruta,
			&estado,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando documento: %w", err)
		}

		d.Tipo = domain.TipoDocumento(tipo)
		d.Estado = domain.EstadoDocumento(estado)
		d.CreatedAt = createdAt
		docs = append(docs, d)
	}

	return docs, rows.Err()
}

func (r *PostgresDocumentoRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM documento WHERE id = $1::uuid`
	cmdTag, err := r.q.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error eliminando documento: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
