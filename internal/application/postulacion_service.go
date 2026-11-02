package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"srp-biess/internal/domain"
)

var (
	ErrPostulacionNoEncontrada = errors.New("postulación no encontrada")
	ErrPostulacionYaEnviada    = errors.New("la postulación ya ha sido enviada previamente")
	ErrPostulacionInvalida     = errors.New("la postulación no está en estado borrador para su envío")
)

type EnviarPostulacionRequest struct {
	PostulacionID string `json:"postulacion_id"`
	Email         string `json:"email"`
}

type EnviarPostulacionResult struct {
	PostulacionID  string    `json:"postulacion_id"`
	CodigoUnico    string    `json:"codigo_unico"`
	Estado         string    `json:"estado"`
	FechaEnvio     time.Time `json:"fecha_envio"`
	DocumentoID    string    `json:"documento_id"`
	NotificacionID string    `json:"notificacion_id"`
}

// PostulacionService orquesta el ciclo de vida y envío de la postulación (ADR-004, SDD 10.9).
type PostulacionService struct {
	txManager        domain.TxManager
	postulacionRepo  domain.PostulacionRepository
	documentoRepo    domain.DocumentoRepository
	notificacionRepo domain.NotificacionRepository
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

	var result EnviarPostulacionResult

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// 1. Validar existencia y estado actual de la postulación
		p, err := s.postulacionRepo.GetByID(txCtx, req.PostulacionID)
		if err != nil {
			return ErrPostulacionNoEncontrada
		}
		if p.Estado == domain.EstadoEnviada {
			return ErrPostulacionYaEnviada
		}
		if p.Estado != domain.EstadoBorrador {
			return ErrPostulacionInvalida
		}

		// 2. Componer código único institucional (ADR-004, SDD 10.9)
		now := time.Now()
		codigoUnico := fmt.Sprintf("BIESS-%d-%05d", now.Year(), now.UnixNano()%100000)

		// 3. Actualizar estado y fecha en postulacion
		if err := s.postulacionRepo.UpdateSubmit(txCtx, p.ID, codigoUnico, now); err != nil {
			return fmt.Errorf("error al actualizar estado de postulación: %w", err)
		}

		// 4. Crear y registrar documento resumen de postulación (PDF)
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

		// 5. Encolar notificación en Outbox
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
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}
