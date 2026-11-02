package domain

import (
	"context"
	"time"
)

// TipoDocumento clasifica los archivos adjuntos requeridos por el concurso (EF-12).
type TipoDocumento string

const (
	DocRespaldoOrganizacional TipoDocumento = "RESPALDO_ORGANIZACIONAL"
	DocTitulo                 TipoDocumento = "TITULO"
	DocExperiencia            TipoDocumento = "EXPERIENCIA"
	DocPublicacion            TipoDocumento = "PUBLICACION"
	DocCapacitacion           TipoDocumento = "CAPACITACION"
	DocFormularioPostulacion  TipoDocumento = "FORMULARIO_POSTULACION"
	DocComprobantePago        TipoDocumento = "COMPROBANTE_PAGO"
)

// EstadoDocumento representa el estado de verificación documental.
type EstadoDocumento string

const (
	EstadoDocPendiente EstadoDocumento = "PENDIENTE"
	EstadoDocValidado  EstadoDocumento = "VALIDADO"
	EstadoDocRechazado EstadoDocumento = "RECHAZADO"
)

// Documento modela los metadatos de un archivo digital respaldado en base de datos (ADR-009).
type Documento struct {
	ID               string          `json:"id"`
	PostulanteID     string          `json:"postulante_id"`
	PostulacionID    string          `json:"postulacion_id,omitempty"`
	Tipo             TipoDocumento   `json:"tipo"`
	NombreOriginal   string          `json:"nombre_original"`
	NombreAlmacenado string          `json:"nombre_almacenado"`
	MimeType         string          `json:"mime_type"`
	TamanoBytes      int64           `json:"tamano_bytes"`
	HashSHA256       string          `json:"hash_sha256"`
	Ruta             string          `json:"ruta"`
	Estado           EstadoDocumento `json:"estado"`
	CreatedAt        time.Time       `json:"created_at"`
}

// DocumentoRepository define el contrato para la persistencia de documentos.
type DocumentoRepository interface {
	Save(ctx context.Context, doc *Documento) error
	GetByID(ctx context.Context, id string) (*Documento, error)
	ListByPostulacionID(ctx context.Context, postulacionID string) ([]Documento, error)
	Delete(ctx context.Context, id string) error
}
