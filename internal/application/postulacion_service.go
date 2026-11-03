package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"srp-biess/internal/domain"
)

var (
	ErrPostulacionNoEncontrada = errors.New("postulación no encontrada")
	ErrPostulacionYaEnviada    = errors.New("la postulación ya ha sido enviada previamente")
	ErrPostulacionInvalida     = errors.New("la postulación no está lista para su envío")
	ErrIdempotencyKeyRequerida = errors.New("Idempotency-Key es requerido")
	ErrIdempotencyKeyConflicto = errors.New("Idempotency-Key ya pertenece a otra postulación")
	ErrConsentimientoRequerido = errors.New("el consentimiento es requerido para enviar")
)

// EnviarPostulacionRequest: PostulanteID sale del JWT; el correo de la notificación se lee del
// registro del postulante, nunca de la petición.
type EnviarPostulacionRequest struct {
	PostulacionID        string `json:"postulacion_id"`
	PostulanteID         string `json:"-"`
	AceptaConsentimiento bool   `json:"aceptaConsentimiento"`
	VersionTerminos      string `json:"versionTerminos"`
	IdempotencyKey       string `json:"-"`
	// Solicitud: IP, navegador y trace_id para la auditoría.
	Solicitud domain.ContextoSolicitud `json:"-"`
}

type EnviarPostulacionResult struct {
	PostulacionID  string    `json:"postulacion_id"`
	CodigoUnico    string    `json:"codigo_unico"`
	Estado         string    `json:"estado"`
	FechaEnvio     time.Time `json:"fecha_envio"`
	DocumentoID    string    `json:"documento_id"`
	NotificacionID string    `json:"notificacion_id"`
	// PDFURL es estable (SDD 7.5.8): una respuesta idempotente puede repetirse días después,
	// cuando una URL firmada de 5 minutos ya habría vencido.
	PDFURL string `json:"pdf_url"`
}

// LectorFormulario lee lo que se imprime en el formulario PDF.
type LectorFormulario interface {
	ObtenerDetalle(ctx context.Context, id, postulanteID string) (*domain.DetalleFormulario, error)
	NombreCargo(ctx context.Context, cargoID int) (string, error)
	TextosConsentimiento(ctx context.Context, postulacionID string) ([]string, error)
	Historial(ctx context.Context, id, postulanteID string) ([]domain.EventoEstado, error)
}

// PostulacionService orquesta el ciclo de vida y envío de la postulación (ADR-004, SDD 10.9).
type PostulacionService struct {
	txManager        domain.TxManager
	postulacionRepo  domain.PostulacionRepository
	documentoRepo    domain.DocumentoRepository
	notificacionRepo domain.NotificacionRepository
	normalizador     domain.NormalizadorPostulacion
	postulantes      domain.PostulanteRepository
	lector           LectorFormulario
	generador        domain.GeneradorFormulario
	almacen          domain.AlmacenDocumentos
	catalogo         domain.CatalogoRepository
	auditoria        domain.AuditoriaRepository
}

// WithAuditoria registra los cambios de estado (SDD 9.9); de ahí sale el historial (SDD 8.5.10).
func (s *PostulacionService) WithAuditoria(auditoria domain.AuditoriaRepository) *PostulacionService {
	s.auditoria = auditoria
	return s
}

// auditarEstado registra la transición en la misma transacción; sin auditoría configurada no hace nada.
func (s *PostulacionService) auditarEstado(ctx context.Context, sol domain.ContextoSolicitud, accion string, p *domain.Postulacion, despues map[string]any) error {
	if s.auditoria == nil {
		return nil
	}
	return s.auditoria.Registrar(ctx, domain.EventoAuditoria{
		ActorID: p.PostulanteID, ActorTipo: domain.ActorPostulante, Accion: accion,
		Entidad: "postulacion", EntidadID: p.ID,
		Antes:   map[string]any{"estado": string(p.Estado)},
		Despues: despues, Solicitud: sol,
	})
}

// Historial devuelve los cambios de estado de la postulación del postulante (SDD 8.5.10).
func (s *PostulacionService) Historial(ctx context.Context, postulacionID, postulanteID string) ([]domain.EventoEstado, error) {
	if s.lector == nil {
		return nil, ErrPersistenciaNoDisponible
	}
	return s.lector.Historial(ctx, postulacionID, postulanteID)
}

// WithCatalogo permite imprimir en el PDF el nombre del país, la provincia y el cantón.
func (s *PostulacionService) WithCatalogo(catalogo domain.CatalogoRepository) *PostulacionService {
	s.catalogo = catalogo
	return s
}

// ubicador consulta el catálogo geográfico (con su caché de Redis, SRP-007) para el PDF.
func (s *PostulacionService) ubicador(ctx context.Context) UbicadorCatalogo {
	if s.catalogo == nil {
		return nil
	}
	return func(tipo, id, padre string) string {
		switch tipo {
		case "pais":
			if paises, err := s.catalogo.GetPaises(ctx); err == nil {
				for _, p := range paises {
					if p.ID == id {
						return p.Nombre
					}
				}
			}
		case "provincia":
			if provincias, err := s.catalogo.GetProvincias(ctx, padre); err == nil {
				for _, p := range provincias {
					if p.ID == id {
						return p.Nombre
					}
				}
			}
		case "canton":
			if cantones, err := s.catalogo.GetCantones(ctx, padre); err == nil {
				for _, c := range cantones {
					if c.ID == id {
						return c.Nombre
					}
				}
			}
		}
		return ""
	}
}

// WithNormalizador configura el volcado JSONB -> tablas dentro de la misma transacción.
func (s *PostulacionService) WithNormalizador(normalizador domain.NormalizadorPostulacion) *PostulacionService {
	s.normalizador = normalizador
	return s
}

// WithFormularioPDF configura lo necesario para generar y guardar el PDF y notificar al correo
// registrado. Sin esto el envío no se confirma: no puede quedar enviada sin formulario.
func (s *PostulacionService) WithFormularioPDF(
	postulantes domain.PostulanteRepository, lector LectorFormulario,
	generador domain.GeneradorFormulario, almacen domain.AlmacenDocumentos,
) *PostulacionService {
	s.postulantes, s.lector, s.generador, s.almacen = postulantes, lector, generador, almacen
	return s
}

func NewPostulacionService(
	txManager domain.TxManager,
	postulacionRepo domain.PostulacionRepository,
	documentoRepo domain.DocumentoRepository,
	notificacionRepo domain.NotificacionRepository,
) *PostulacionService {
	return &PostulacionService{
		txManager:        txManager,
		postulacionRepo:  postulacionRepo,
		documentoRepo:    documentoRepo,
		notificacionRepo: notificacionRepo,
	}
}

// URLFormularioPDF es la ruta estable de descarga del formulario (SDD 7.5.8).
func URLFormularioPDF(postulacionID string) string {
	return "/api/v1/postulaciones/" + postulacionID + "/pdf"
}

// EnviarPostulacion ejecuta la transacción crítica atómica multi-tabla (SRP-005, ADR-004, SDD 10.9):
// estado ENVIADA, código único, PDF generado y guardado, consentimiento, notificación e idempotencia.
// Si el PDF no se genera o no se guarda, no cambia nada (SRS 3.5.20).
func (s *PostulacionService) EnviarPostulacion(ctx context.Context, req EnviarPostulacionRequest) (*EnviarPostulacionResult, error) {
	if req.PostulacionID == "" {
		return nil, errors.New("el ID de postulación es requerido")
	}
	if req.PostulanteID == "" {
		return nil, domain.ErrUnauthorized
	}
	if req.IdempotencyKey == "" {
		return nil, ErrIdempotencyKeyRequerida
	}
	if !req.AceptaConsentimiento {
		return nil, ErrConsentimientoRequerido
	}
	if strings.TrimSpace(req.VersionTerminos) != "" && !versionTerminosValida(req.VersionTerminos) {
		return nil, ErrVersionTerminosInvalida
	}
	if s.postulantes == nil || s.lector == nil || s.generador == nil || s.almacen == nil {
		return nil, ErrPersistenciaNoDisponible
	}

	var result EnviarPostulacionResult
	pendiente, definitivo := "", "" // temporal sin confirmar y archivo renombrado
	defer func() {
		if pendiente != "" {
			_ = s.almacen.Descartar(pendiente)
		}
		if definitivo != "" {
			_ = s.almacen.Eliminar(definitivo)
		}
	}()

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// El bloqueo serializa envíos concurrentes de la misma postulación.
		p, err := s.postulacionRepo.GetByIDForUpdate(txCtx, req.PostulacionID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return ErrPostulacionNoEncontrada
			}
			return fmt.Errorf("error al consultar postulación: %w", err)
		}
		// La propiedad se valida ANTES de responder desde la idempotencia: con el UUID y la llave
		// de otra persona no se obtiene su respuesta guardada.
		if p.PostulanteID != req.PostulanteID {
			return domain.ErrAccesoDenegado
		}
		if record, lookupErr := s.postulacionRepo.GetIdempotentSubmission(txCtx, req.IdempotencyKey); lookupErr == nil {
			if record.PostulacionID != p.ID {
				return ErrIdempotencyKeyConflicto
			}
			return json.Unmarshal(record.Response, &result)
		} else if !errors.Is(lookupErr, domain.ErrNotFound) {
			return fmt.Errorf("error consultando idempotencia de envío: %w", lookupErr)
		}
		if p.Estado == domain.EstadoEnviada {
			return ErrPostulacionYaEnviada
		}
		if p.Estado != domain.EstadoListaParaEnvio {
			return ErrPostulacionInvalida
		}

		if s.normalizador != nil {
			if err := s.normalizador.Normalizar(txCtx, p.ID); err != nil {
				return fmt.Errorf("error normalizando datos de postulación: %w", err)
			}
		}

		now := time.Now()
		codigoUnico, err := s.postulacionRepo.NextCodigoUnico(txCtx)
		if err != nil {
			return fmt.Errorf("error al generar código correlativo único: %w", err)
		}
		if err := s.postulacionRepo.UpdateSubmit(txCtx, p.ID, codigoUnico, now); err != nil {
			if errors.Is(err, domain.ErrPostulacionYaEnviada) {
				return ErrPostulacionYaEnviada
			}
			return fmt.Errorf("error al actualizar estado de postulación: %w", err)
		}

		// Correo de la notificación: el registrado por el postulante (RF-EF12-19).
		postulante, err := s.postulantes.GetByID(txCtx, p.PostulanteID)
		if err != nil {
			return fmt.Errorf("error consultando el postulante: %w", err)
		}

		// Formulario PDF real: se genera, se guarda en el almacén y se registra con su hash.
		doc, err := s.generarFormulario(txCtx, p, codigoUnico, now, req.VersionTerminos, &pendiente, &definitivo)
		if err != nil {
			return err
		}
		if err := s.documentoRepo.Save(txCtx, doc); err != nil {
			return fmt.Errorf("error al registrar documento de formulario: %w", err)
		}

		consentVersion := req.VersionTerminos
		if consentVersion == "" {
			consentVersion = "legacy"
		}
		if err := s.postulacionRepo.SaveConsentimiento(txCtx, p.ID, consentVersion); err != nil {
			return fmt.Errorf("error al registrar consentimiento: %w", err)
		}

		notif := &domain.Notificacion{
			PostulacionID: p.ID,
			PostulanteID:  p.PostulanteID,
			Tipo:          "CONFIRMACION_POSTULACION",
			Destinatario:  postulante.Email,
			Asunto:        fmt.Sprintf("Confirmación de Postulación BIESS — %s", codigoUnico),
			Plantilla:     "confirmacion_postulacion.html",
			Payload: map[string]any{
				"codigo_unico": codigoUnico,
				"fecha_envio":  now.Format(time.RFC3339),
				"cargo_id":     p.CargoID,
				// El remitente adjunta el formulario guardado (SRS 2: "Enviar correo con PDF").
				"documento_id": doc.ID,
			},
		}
		if err := s.notificacionRepo.Enqueue(txCtx, notif); err != nil {
			return fmt.Errorf("error al encolar notificación: %w", err)
		}

		result = EnviarPostulacionResult{
			PostulacionID:  p.ID,
			CodigoUnico:    codigoUnico,
			Estado:         string(domain.EstadoEnviada),
			FechaEnvio:     now,
			DocumentoID:    doc.ID,
			NotificacionID: notif.ID,
			PDFURL:         URLFormularioPDF(p.ID),
		}
		if err := s.auditarEstado(txCtx, req.Solicitud, domain.AccionPostulacionEnviada, p,
			map[string]any{"estado": string(domain.EstadoEnviada), "codigoUnico": codigoUnico}); err != nil {
			return fmt.Errorf("error auditando el envío: %w", err)
		}
		payload, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("error serializando respuesta idempotente: %w", err)
		}
		return s.postulacionRepo.SaveIdempotentSubmission(txCtx, req.IdempotencyKey, p.ID, payload)
	})
	if err != nil {
		return nil, err
	}
	definitivo = "" // COMMIT confirmado: el formulario se conserva
	return &result, nil
}

// generarFormulario renderiza y guarda el PDF; deja en pendiente/definitivo lo que hay que limpiar
// si la transacción no confirma.
func (s *PostulacionService) generarFormulario(
	ctx context.Context, p *domain.Postulacion, codigo string, fecha time.Time, versionTerminos string,
	pendiente, definitivo *string,
) (*domain.Documento, error) {
	detalle, err := s.lector.ObtenerDetalle(ctx, p.ID, p.PostulanteID)
	if err != nil {
		return nil, fmt.Errorf("error leyendo la postulación para el formulario: %w", err)
	}
	cargo := ""
	if id, convErr := strconv.Atoi(p.CargoID); convErr == nil && id > 0 {
		if cargo, err = s.lector.NombreCargo(ctx, id); err != nil {
			return nil, fmt.Errorf("error leyendo el cargo: %w", err)
		}
	}
	textos, err := s.lector.TextosConsentimiento(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	formulario := ArmarFormularioPDF(detalle, codigo, fecha, cargo, versionTerminos, s.ubicador(ctx))
	if len(textos) > 0 {
		formulario.TextoConsentimiento = TextoConsentimientos(textos)
	}
	contenido, err := s.generador.Generar(formulario)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFormularioNoGenerado, err)
	}
	if !domain.EsPDF(contenido) {
		return nil, domain.ErrFormularioNoGenerado
	}

	ref, tamano, err := s.almacen.EscribirTemporal(bytes.NewReader(contenido))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFormularioNoGenerado, err)
	}
	*pendiente = ref
	nombre := uuid.NewString() + ".pdf"
	ruta, err := s.almacen.Confirmar(ref, nombre)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFormularioNoGenerado, err)
	}
	*pendiente, *definitivo = "", nombre

	suma := sha256.Sum256(contenido)
	return &domain.Documento{
		PostulanteID:     p.PostulanteID,
		PostulacionID:    p.ID,
		Tipo:             domain.DocFormularioPostulacion,
		NombreOriginal:   domain.NombreFormularioPDF(codigo),
		NombreAlmacenado: nombre,
		MimeType:         "application/pdf",
		TamanoBytes:      tamano,
		HashSHA256:       hex.EncodeToString(suma[:]),
		Ruta:             ruta,
		Estado:           domain.EstadoDocValidado,
	}, nil
}

// DescargarFormulario devuelve el PDF generado al enviar (SDD 7.5.8: valida pertenencia).
func (s *PostulacionService) DescargarFormulario(ctx context.Context, postulacionID, postulanteID string) (*domain.Documento, io.ReadCloser, error) {
	if s.almacen == nil {
		return nil, nil, ErrPersistenciaNoDisponible
	}
	p, err := s.postulacionRepo.GetByID(ctx, postulacionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil, ErrPostulacionNoEncontrada
		}
		return nil, nil, err
	}
	if p.PostulanteID != postulanteID {
		return nil, nil, domain.ErrAccesoDenegado
	}
	docs, err := s.documentoRepo.ListByPostulacionID(ctx, p.ID)
	if err != nil {
		return nil, nil, err
	}
	for i := range docs {
		if docs[i].Tipo == domain.DocFormularioPostulacion {
			archivo, err := s.almacen.Abrir(docs[i].NombreAlmacenado)
			if err != nil {
				return nil, nil, err
			}
			return &docs[i], archivo, nil
		}
	}
	return nil, nil, domain.ErrDocumentoNoEncontrado
}
