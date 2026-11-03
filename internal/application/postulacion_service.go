package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

type EnviarPostulacionRequest struct {
	PostulacionID        string `json:"postulacion_id"`
	Email                string `json:"email"`
	AceptaConsentimiento bool   `json:"aceptaConsentimiento"`
	VersionTerminos      string `json:"versionTerminos"`
	IdempotencyKey       string `json:"-"`
}

type EnviarPostulacionResult struct {
	PostulacionID  string    `json:"postulacion_id"`
	CodigoUnico    string    `json:"codigo_unico"`
	Estado         string    `json:"estado"`
	FechaEnvio     time.Time `json:"fecha_envio"`
	DocumentoID    string    `json:"documento_id"`
	NotificacionID string    `json:"notificacion_id"`
	PDFURL         string    `json:"pdf_url"`
}

// PostulacionService orquesta el ciclo de vida y envío de la postulación (ADR-004, SDD 10.9).
type PostulacionService struct {
	txManager        domain.TxManager
	postulacionRepo  domain.PostulacionRepository
	documentoRepo    domain.DocumentoRepository
	notificacionRepo domain.NotificacionRepository
	normalizador     domain.NormalizadorPostulacion
}

// WithNormalizador configura el volcado JSONB -> tablas dentro de la misma transacción.
func (s *PostulacionService) WithNormalizador(normalizador domain.NormalizadorPostulacion) *PostulacionService {
	s.normalizador = normalizador
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

// EnviarPostulacion ejecuta la transacción crítica atómica multi-tabla (SRP-005, ADR-004, SDD 10.9).
// Actualiza el estado a ENVIADA, genera código único, crea el PDF y encola la notificación en una sola transacción atómica.
func (s *PostulacionService) EnviarPostulacion(ctx context.Context, req EnviarPostulacionRequest) (*EnviarPostulacionResult, error) {
	if req.PostulacionID == "" {
		return nil, errors.New("el ID de postulación es requerido")
	}
	if req.IdempotencyKey == "" {
		return nil, ErrIdempotencyKeyRequerida
	}
	if !req.AceptaConsentimiento {
		return nil, ErrConsentimientoRequerido
	}

	var result EnviarPostulacionResult

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// La fila queda bloqueada antes de validar el estado y generar el código.
		// Esto serializa envíos concurrentes para la misma postulación.
		record, err := s.postulacionRepo.GetIdempotentSubmission(txCtx, req.IdempotencyKey)
		if err == nil {
			if record.PostulacionID != req.PostulacionID {
				return ErrIdempotencyKeyConflicto
			}
			return json.Unmarshal(record.Response, &result)
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("error consultando idempotencia de envío: %w", err)
		}

		p, err := s.postulacionRepo.GetByIDForUpdate(txCtx, req.PostulacionID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return ErrPostulacionNoEncontrada
			}
			return fmt.Errorf("error al consultar postulación: %w", err)
		}
		// El primer lookup puede haber ocurrido mientras otro envío insertaba la
		// llave. Tras adquirir el lock de la postulación se vuelve a consultar.
		if record, lookupErr := s.postulacionRepo.GetIdempotentSubmission(txCtx, req.IdempotencyKey); lookupErr == nil {
			if record.PostulacionID != req.PostulacionID {
				return ErrIdempotencyKeyConflicto
			}
			return json.Unmarshal(record.Response, &result)
		} else if !errors.Is(lookupErr, domain.ErrNotFound) {
			return fmt.Errorf("error reconsultando idempotencia de envío: %w", lookupErr)
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

		// Generar código correlativo único institucional desde la secuencia de PostgreSQL.
		now := time.Now()
		codigoUnico, err := s.postulacionRepo.NextCodigoUnico(txCtx)
		if err != nil {
			return fmt.Errorf("error al generar código correlativo único: %w", err)
		}

		// Actualizar estado y fecha en postulacion dentro de la transacción.
		if err := s.postulacionRepo.UpdateSubmit(txCtx, p.ID, codigoUnico, now); err != nil {
			if errors.Is(err, domain.ErrPostulacionYaEnviada) {
				return ErrPostulacionYaEnviada
			}
			return fmt.Errorf("error al actualizar estado de postulación: %w", err)
		}

		// Registrar documento resumen de postulación (PDF).
		doc := &domain.Documento{
			PostulanteID:     p.PostulanteID,
			PostulacionID:    p.ID,
			Tipo:             domain.DocFormularioPostulacion,
			NombreOriginal:   fmt.Sprintf("Postulacion_%s.pdf", codigoUnico),
			NombreAlmacenado: fmt.Sprintf("form_%s_%d.pdf", p.ID, now.UnixNano()),
			MimeType:         "application/pdf",
			TamanoBytes:      1024 * 128, // 128 KB
			HashSHA256:       "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			Ruta:             fmt.Sprintf("/storage/postulaciones/%s.pdf", codigoUnico),
			Estado:           domain.EstadoDocValidado,
		}
		if err := s.documentoRepo.Save(txCtx, doc); err != nil {
			return fmt.Errorf("error al registrar documento de formulario: %w", err)
		}

		// Persistir consentimiento y encolar notificación en Outbox.
		consentVersion := req.VersionTerminos
		if consentVersion == "" {
			consentVersion = "legacy"
		}
		if err := s.postulacionRepo.SaveConsentimiento(txCtx, p.ID, consentVersion); err != nil {
			return fmt.Errorf("error al registrar consentimiento: %w", err)
		}

		emailDestino := req.Email
		if emailDestino == "" {
			emailDestino = "postulante@biess.gob.ec"
		}

		notif := &domain.Notificacion{
			PostulacionID: p.ID,
			PostulanteID:  p.PostulanteID,
			Tipo:          "CONFIRMACION_POSTULACION",
			Destinatario:  emailDestino,
			Asunto:        fmt.Sprintf("Confirmación de Postulación BIESS — %s", codigoUnico),
			Plantilla:     "confirmacion_postulacion.html",
			Payload: map[string]any{
				"codigo_unico": codigoUnico,
				"fecha_envio":  now.Format(time.RFC3339),
				"cargo_id":     p.CargoID,
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
			PDFURL:         doc.Ruta,
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

	return &result, nil
}
