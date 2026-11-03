package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"srp-biess/internal/domain"
)

// DefaultVigenciaDescarga es la vigencia de la URL firmada (SDD 9.7: 5 min).
const DefaultVigenciaDescarga = 5 * time.Minute

// DocumentoService carga, elimina y sirve documentos de la postulación (SRP-017b, SDD 7.5.5, 9.7).
type DocumentoService struct {
	tx        domain.TxManager
	repo      domain.CargaDocumentoRepository
	almacen   domain.AlmacenDocumentos
	firmador  domain.FirmadorDescarga
	auditoria domain.AuditoriaRepository
	limites   domain.LimitesDocumento
	vigencia  time.Duration
	ahora     func() time.Time
}

func NewDocumentoService(
	tx domain.TxManager, repo domain.CargaDocumentoRepository, almacen domain.AlmacenDocumentos,
	firmador domain.FirmadorDescarga, auditoria domain.AuditoriaRepository,
	limites domain.LimitesDocumento, vigencia time.Duration,
) *DocumentoService {
	if vigencia <= 0 {
		vigencia = DefaultVigenciaDescarga
	}
	return &DocumentoService{
		tx: tx, repo: repo, almacen: almacen, firmador: firmador, auditoria: auditoria,
		limites: limites, vigencia: vigencia, ahora: time.Now,
	}
}

// WithReloj permite fijar la hora en pruebas.
func (s *DocumentoService) WithReloj(ahora func() time.Time) *DocumentoService {
	s.ahora = ahora
	return s
}

// DocumentoCargado es la respuesta de la carga: metadatos y URL firmada vigente.
type DocumentoCargado struct {
	Documento   *domain.Documento
	URLDescarga string
}

func (s *DocumentoService) disponible() error {
	if s.tx == nil || s.repo == nil || s.almacen == nil || s.firmador == nil || s.auditoria == nil {
		return ErrPersistenciaNoDisponible
	}
	return nil
}

// Subir valida y guarda un PDF. Orden (sin huérfanos): temporal → BEGIN → bloqueo y cuota →
// rename → INSERT → auditoría → COMMIT. El rename va antes del INSERT porque la fila guarda la
// ruta definitiva; si el INSERT, la auditoría o el COMMIT fallan, se borra el archivo definitivo.
func (s *DocumentoService) Subir(
	ctx context.Context, sol domain.ContextoSolicitud, postulacionID, postulanteID string,
	tipo domain.TipoDocumento, nombreOriginal string, contenido io.Reader,
) (*DocumentoCargado, error) {
	if err := s.disponible(); err != nil {
		return nil, err
	}
	if !tipo.CargablePorPostulante() {
		return nil, domain.ErrDocumentoTipoNoCargable
	}
	nombre := SanearNombreArchivo(nombreOriginal)
	if !strings.HasSuffix(strings.ToLower(nombre), ".pdf") {
		return nil, domain.ErrDocumentoNoPDF // SDD 9.7: extensión .pdf además de la firma
	}

	// Firma %PDF- en los primeros bytes; luego una sola pasada: límite, SHA-256 y temporal.
	cabecera := make([]byte, len(domain.FirmaPDF))
	leidos, _ := io.ReadFull(contenido, cabecera)
	if !domain.EsPDF(cabecera[:leidos]) {
		return nil, domain.ErrDocumentoNoPDF
	}
	limite := s.limites.MaximoPara(tipo)
	hash := sha256.New()
	flujo := io.TeeReader(io.LimitReader(io.MultiReader(bytes.NewReader(cabecera), contenido), limite+1), hash)
	ref, tamano, err := s.almacen.EscribirTemporal(flujo)
	if err != nil {
		return nil, err
	}
	pendiente := ref // temporal sin confirmar: se descarta al salir
	definitivo := "" // archivo renombrado: se borra si la transacción no confirma
	defer func() {
		if pendiente != "" {
			_ = s.almacen.Descartar(pendiente)
		}
		if definitivo != "" {
			_ = s.almacen.Eliminar(definitivo)
		}
	}()
	if tamano > limite {
		return nil, domain.ErrDocumentoExcedeTamano
	}

	doc := &domain.Documento{
		PostulanteID: postulanteID, PostulacionID: postulacionID, Tipo: tipo,
		NombreOriginal: nombre, NombreAlmacenado: uuid.NewString() + ".pdf", MimeType: "application/pdf",
		TamanoBytes: tamano, HashSHA256: hex.EncodeToString(hash.Sum(nil)), Estado: domain.EstadoDocPendiente,
	}
	err = s.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		estado, err := s.repo.BloquearPostulacion(txCtx, postulacionID, postulanteID)
		if err != nil {
			return err
		}
		if estado != "BORRADOR" {
			return domain.ErrPostulacionCerrada
		}
		total, err := s.repo.TotalBytes(txCtx, postulacionID)
		if err != nil {
			return err
		}
		if total+tamano > s.limites.TotalBytes {
			return domain.ErrDocumentoCuotaExcedida
		}
		ruta, err := s.almacen.Confirmar(ref, doc.NombreAlmacenado)
		if err != nil {
			return err
		}
		pendiente, definitivo, doc.Ruta = "", doc.NombreAlmacenado, ruta
		if err := s.repo.Insertar(txCtx, doc); err != nil {
			return err
		}
		return s.auditoria.Registrar(txCtx, eventoDocumento(sol, domain.AccionDocumentoCargado, postulanteID, doc))
	})
	if errors.Is(err, domain.ErrAccesoDenegado) {
		s.auditarAccesoDenegado(ctx, sol, postulacionID, postulanteID, "POST /postulaciones/{id}/documentos")
	}
	if err != nil {
		return nil, err
	}
	definitivo = "" // COMMIT confirmado: el archivo se conserva
	return &DocumentoCargado{Documento: doc, URLDescarga: s.urlDescarga(postulacionID, doc.ID)}, nil
}

// Eliminar borra metadatos y archivo. Orden: DELETE → auditoría → COMMIT → borrar archivo; un
// archivo huérfano es inofensivo, una fila sin archivo rompería la descarga.
func (s *DocumentoService) Eliminar(ctx context.Context, sol domain.ContextoSolicitud, postulacionID, postulanteID, docID string) error {
	if err := s.disponible(); err != nil {
		return err
	}
	var doc *domain.Documento
	err := s.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		estado, err := s.repo.BloquearPostulacion(txCtx, postulacionID, postulanteID)
		if err != nil {
			return err
		}
		if estado != "BORRADOR" {
			return domain.ErrPostulacionCerrada
		}
		if doc, err = s.repo.ObtenerDePostulacion(txCtx, docID, postulacionID); err != nil {
			return err
		}
		if err := s.repo.EliminarDePostulacion(txCtx, docID, postulacionID); err != nil {
			return err
		}
		return s.auditoria.Registrar(txCtx, eventoDocumento(sol, domain.AccionDocumentoEliminado, postulanteID, doc))
	})
	if errors.Is(err, domain.ErrAccesoDenegado) {
		s.auditarAccesoDenegado(ctx, sol, postulacionID, postulanteID, "DELETE /postulaciones/{id}/documentos/{docId}")
	}
	if err != nil {
		return err
	}
	if err := s.almacen.Eliminar(doc.NombreAlmacenado); err != nil {
		log.Printf("[WARN] documento %s eliminado de la base, pero no del disco: %v", doc.ID, err)
	}
	return nil
}

// Metadatos devuelve el documento con una URL firmada nueva (la anterior pudo vencer).
func (s *DocumentoService) Metadatos(ctx context.Context, postulacionID, postulanteID, docID string) (*DocumentoCargado, error) {
	if err := s.disponible(); err != nil {
		return nil, err
	}
	doc, err := s.repo.ObtenerDePostulante(ctx, docID, postulanteID)
	if err != nil {
		return nil, err
	}
	if doc.PostulacionID != postulacionID {
		return nil, domain.ErrDocumentoNoEncontrado
	}
	return &DocumentoCargado{Documento: doc, URLDescarga: s.urlDescarga(postulacionID, doc.ID)}, nil
}

// Descargar exige firma vigente (403) y que el documento sea del postulante del JWT (404).
func (s *DocumentoService) Descargar(
	ctx context.Context, postulacionID, postulanteID, docID string, vence int64, firma string,
) (*domain.Documento, io.ReadCloser, error) {
	if err := s.disponible(); err != nil {
		return nil, nil, err
	}
	if err := s.firmador.Verificar(docID, vence, firma, s.ahora()); err != nil {
		return nil, nil, err
	}
	doc, err := s.repo.ObtenerDePostulante(ctx, docID, postulanteID)
	if err != nil {
		return nil, nil, err
	}
	if doc.PostulacionID != postulacionID {
		return nil, nil, domain.ErrDocumentoNoEncontrado
	}
	archivo, err := s.almacen.Abrir(doc.NombreAlmacenado)
	if err != nil {
		return nil, nil, err
	}
	return doc, archivo, nil
}

func (s *DocumentoService) urlDescarga(postulacionID, docID string) string {
	vence := s.ahora().Add(s.vigencia)
	return "/api/v1/postulaciones/" + postulacionID + "/documentos/" + docID + "/download?exp=" +
		strconv.FormatInt(vence.Unix(), 10) + "&sig=" + s.firmador.Firmar(docID, vence)
}

// SanearNombreArchivo deja solo el nombre base, sin caracteres de control, en 255 caracteres
// como máximo. Nunca se usa en la ruta en disco (SDD 9.7).
func SanearNombreArchivo(nombre string) string {
	nombre = filepath.Base(strings.ReplaceAll(nombre, "\\", "/"))
	nombre = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, nombre)
	nombre = strings.TrimSpace(nombre)
	if nombre == "." || nombre == "/" || nombre == "" {
		nombre = "documento.pdf"
	}
	if r := []rune(nombre); len(r) > 255 {
		nombre = string(r[len(r)-255:]) // conserva la extensión
	}
	return nombre
}

func eventoDocumento(sol domain.ContextoSolicitud, accion, postulanteID string, d *domain.Documento) domain.EventoAuditoria {
	datos := map[string]any{
		"postulacionId": d.PostulacionID, "tipo": d.Tipo, "nombreOriginal": d.NombreOriginal,
		"tamanoBytes": d.TamanoBytes, "hashSha256": d.HashSHA256,
	}
	e := domain.EventoAuditoria{
		ActorID: postulanteID, ActorTipo: domain.ActorPostulante, Accion: accion,
		Entidad: "documento", EntidadID: d.ID, Solicitud: sol,
	}
	if accion == domain.AccionDocumentoEliminado {
		e.Antes = datos
	} else {
		e.Despues = datos
	}
	return e
}

func (s *DocumentoService) auditarAccesoDenegado(ctx context.Context, sol domain.ContextoSolicitud, postulacionID, postulanteID, operacion string) {
	err := s.auditoria.Registrar(ctx, domain.EventoAuditoria{
		ActorID: postulanteID, ActorTipo: domain.ActorPostulante, Accion: domain.AccionAccesoDenegado,
		Entidad: "postulacion", EntidadID: postulacionID, Despues: map[string]any{"operacion": operacion}, Solicitud: sol,
	})
	if err != nil {
		log.Printf("[WARN] no se pudo auditar el acceso denegado a la postulación %s: %v", postulacionID, err)
	}
}
