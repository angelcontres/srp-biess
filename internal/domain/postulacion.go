package domain

import (
	"context"
	"errors"
	"time"
)

// Errores de dominio para postulación
var (
	ErrPostulacionYaEnviada = errors.New("la postulación ya fue enviada previamente")
)

// EstadoPostulacion representa las fases de la postulación (BORRADOR, LISTA_PARA_ENVIO, ENVIADA, APROBADA, RECHAZADA)
type EstadoPostulacion string

const (
	EstadoBorrador       EstadoPostulacion = "BORRADOR"
	EstadoListaParaEnvio EstadoPostulacion = "LISTA_PARA_ENVIO"
	EstadoEnviada        EstadoPostulacion = "ENVIADA"
	EstadoAprobada       EstadoPostulacion = "APROBADA"
	EstadoRechazada      EstadoPostulacion = "RECHAZADA"
)

// Postulacion representa la solicitud formal a un cargo del concurso.
type Postulacion struct {
	ID                string            `json:"id"`
	CodigoPostulacion string            `json:"codigo_postulacion"` // Ej. BIESS-2026-00001
	PostulanteID      string            `json:"postulante_id"`
	CargoID           string            `json:"cargo_id"`
	Estado            EstadoPostulacion `json:"estado"`
	ExperienciaMeses  int               `json:"experiencia_meses"`
	DeclaracionJurada bool              `json:"declaracion_jurada"`
	CertificadoPDFURL string            `json:"certificado_pdf_url,omitempty"`
	FechaEnvio        *time.Time        `json:"fecha_envio,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// PostulacionRepository define el puerto para persistir postulaciones.
type PostulacionRepository interface {
	GetByID(ctx context.Context, id string) (*Postulacion, error)
	GetByIDForUpdate(ctx context.Context, id string) (*Postulacion, error)
	GetByPostulanteID(ctx context.Context, postulanteID string) (*Postulacion, error)
	Save(ctx context.Context, p *Postulacion) error
	UpdateEstado(ctx context.Context, id string, estado EstadoPostulacion) error
	UpdateSubmit(ctx context.Context, id, codigoUnico string, fechaEnvio time.Time) error
	NextCodigoUnico(ctx context.Context) (string, error)
	GetIdempotentSubmission(ctx context.Context, key string) (*IdempotentSubmission, error)
	SaveIdempotentSubmission(ctx context.Context, key, postulacionID string, response []byte) error
	SaveConsentimiento(ctx context.Context, postulacionID, version string) error
	// SaveTextosConsentimiento guarda los textos que el postulante aceptó (consentimiento.textos_aceptados).
	SaveTextosConsentimiento(ctx context.Context, postulacionID string, textos []string) error
}

// IdempotentSubmission contiene la respuesta persistida de un envío ya procesado.
type IdempotentSubmission struct {
	Key           string
	PostulacionID string
	Response      []byte
}
