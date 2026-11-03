package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"srp-biess/internal/domain"
)

// PostgresDocumentoRepository también implementa domain.CargaDocumentoRepository (SRP-017b).
// Todas las sentencias usan la transacción del contexto si existe (TxManager).
var _ domain.CargaDocumentoRepository = (*PostgresDocumentoRepository)(nil)

const columnasDocumento = `id, postulante_id, COALESCE(postulacion_id::text, ''), tipo::text, nombre_original,
	nombre_almacenado, mime_type, tamano_bytes, hash_sha256, ruta, estado::text, created_at`

func (r *PostgresDocumentoRepository) BloquearPostulacion(ctx context.Context, postulacionID, postulanteID string) (string, error) {
	if !idsValidos(postulacionID, postulanteID) {
		return "", domain.ErrNotFound
	}
	q := r.getQuerier(ctx)
	var estado, duenio string
	err := q.QueryRow(ctx,
		`SELECT estado::text, postulante_id::text FROM postulacion WHERE id = $1::uuid FOR UPDATE`,
		postulacionID).Scan(&estado, &duenio)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return "", domain.ErrNotFound
	case err != nil:
		return "", fmt.Errorf("error bloqueando la postulación: %w", err)
	case duenio != postulanteID:
		return "", domain.ErrAccesoDenegado
	}
	return estado, nil
}

func (r *PostgresDocumentoRepository) TotalBytes(ctx context.Context, postulacionID string) (int64, error) {
	var total int64
	err := r.getQuerier(ctx).QueryRow(ctx,
		`SELECT COALESCE(SUM(tamano_bytes), 0) FROM documento WHERE postulacion_id = $1::uuid`,
		postulacionID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("error sumando documentos: %w", err)
	}
	return total, nil
}

// Insertar traduce el rechazo de ck_documento_tamano (23514) a un error de dominio: el handler
// responde 413 y nunca 500.
func (r *PostgresDocumentoRepository) Insertar(ctx context.Context, d *domain.Documento) error {
	err := r.Save(ctx, d)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23514" && pgErr.ConstraintName == "ck_documento_tamano" {
		return domain.ErrDocumentoExcedeTamano
	}
	if err != nil {
		return fmt.Errorf("error registrando documento: %w", err)
	}
	return nil
}

func (r *PostgresDocumentoRepository) ObtenerDePostulacion(ctx context.Context, docID, postulacionID string) (*domain.Documento, error) {
	if !idsValidos(docID, postulacionID) {
		return nil, domain.ErrDocumentoNoEncontrado
	}
	return r.leerDocumento(ctx, `WHERE id = $1::uuid AND postulacion_id = $2::uuid`, docID, postulacionID)
}

func (r *PostgresDocumentoRepository) ObtenerDePostulante(ctx context.Context, docID, postulanteID string) (*domain.Documento, error) {
	if !idsValidos(docID, postulanteID) {
		return nil, domain.ErrDocumentoNoEncontrado
	}
	return r.leerDocumento(ctx, `WHERE id = $1::uuid AND postulante_id = $2::uuid`, docID, postulanteID)
}

func (r *PostgresDocumentoRepository) EliminarDePostulacion(ctx context.Context, docID, postulacionID string) error {
	if !idsValidos(docID, postulacionID) {
		return domain.ErrDocumentoNoEncontrado
	}
	tag, err := r.getQuerier(ctx).Exec(ctx,
		`DELETE FROM documento WHERE id = $1::uuid AND postulacion_id = $2::uuid`, docID, postulacionID)
	if err != nil {
		return fmt.Errorf("error eliminando documento: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrDocumentoNoEncontrado
	}
	return nil
}

func (r *PostgresDocumentoRepository) leerDocumento(ctx context.Context, where string, args ...any) (*domain.Documento, error) {
	var d domain.Documento
	var tipo, estado string
	err := r.getQuerier(ctx).QueryRow(ctx, `SELECT `+columnasDocumento+` FROM documento `+where, args...).Scan(
		&d.ID, &d.PostulanteID, &d.PostulacionID, &tipo, &d.NombreOriginal, &d.NombreAlmacenado,
		&d.MimeType, &d.TamanoBytes, &d.HashSHA256, &d.Ruta, &estado, &d.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDocumentoNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error consultando documento: %w", err)
	}
	d.Tipo = domain.TipoDocumento(tipo)
	d.Estado = domain.EstadoDocumento(estado)
	return &d, nil
}
