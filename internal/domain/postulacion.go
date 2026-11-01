package domain

import (
	"context"
	"time"
)

// EstadoPostulacion representa las fases de la postulación (BORRADOR, ENVIADA, APROBADA, RECHAZADA)
type EstadoPostulacion string

const (
	EstadoBorrador  EstadoPostulacion = "BORRADOR"
	EstadoEnviada   EstadoPostulacion = "ENVIADA"
	EstadoAprobada  EstadoPostulacion = "APROBADA"
	EstadoRechazada EstadoPostulacion = "RECHAZADA"
)

// Postulacion representa la solicitud formal a un cargo del concurso.
type Postulacion struct {
	ID                 string            `json:"id"`
	CodigoPostulacion  string            `json:"codigo_postulacion"` // Ej. BIESS-2026-0001
	PostulanteID       string            `json:"postulante_id"`
	CargoID            string            `json:"cargo_id"`
	Estado             EstadoPostulacion `json:"estado"`
	ExperienciaMeses   int               `json:"experiencia_meses"`
	DeclaracionJurada  bool              `json:"declaracion_jurada"`
	ComprobantePagoURL string            `json:"comprobante_pago_url,omitempty"`
	CertificadoPDFURL  string            `json:"certificado_pdf_url,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// PostulacionRepository define el puerto para persistir postulaciones.
type PostulacionRepository interface {
	GetByID(ctx context.Context, id string) (*Postulacion, error)
	GetByPostulanteID(ctx context.Context, postulanteID string) (*Postulacion, error)
	Save(ctx context.Context, p *Postulacion) error
	UpdateEstado(ctx context.Context, id string, estado EstadoPostulacion) error
}
