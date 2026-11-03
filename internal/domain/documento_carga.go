package domain

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"
)

// Carga, borrado y descarga de documentos (SRP-017b, SDD 7.5.5, SDD 9.7).

const mb = 1024 * 1024

// LimitesDocumento son los topes de tamaño (SDD 16.4.4). La base solo garantiza 4 MB por
// documento (CHECK ck_documento_tamano); el tope de 2 MB y el acumulado son reglas de aplicación.
type LimitesDocumento struct {
	RespaldoBytes int64 // MAX_ORGANIZATIONAL_SUPPORT_MB
	OtrosBytes    int64 // MAX_DOCUMENT_MB
	TotalBytes    int64 // MAX_TOTAL_UPLOAD_MB
}

// LimitesPorDefecto son los valores de SDD 16 (4 MB, 2 MB y 50 MB).
var LimitesPorDefecto = LimitesDocumento{RespaldoBytes: 4 * mb, OtrosBytes: 2 * mb, TotalBytes: 50 * mb}

// MaximoPara devuelve el tope por documento según su tipo.
func (l LimitesDocumento) MaximoPara(t TipoDocumento) int64 {
	if t == DocRespaldoOrganizacional {
		return l.RespaldoBytes
	}
	return l.OtrosBytes
}

// CargablePorPostulante indica si el postulante puede subir este tipo. FORMULARIO_POSTULACION
// lo genera el sistema al enviar (#43).
func (t TipoDocumento) CargablePorPostulante() bool {
	switch t {
	case DocRespaldoOrganizacional, DocTitulo, DocExperiencia, DocPublicacion, DocCapacitacion:
		return true
	}
	return false
}

// FirmaPDF son los magic bytes de un PDF (`%PDF-`).
var FirmaPDF = []byte("%PDF-")

// EsPDF valida el tipo real por los primeros bytes, no por la extensión ni el Content-Type.
func EsPDF(cabecera []byte) bool {
	return bytes.HasPrefix(cabecera, FirmaPDF)
}

var (
	ErrDocumentoTipoNoCargable = errors.New("el tipo de documento no se puede cargar")
	ErrDocumentoNoPDF          = errors.New("el archivo no es un PDF")
	ErrDocumentoExcedeTamano   = errors.New("el documento supera el tamaño permitido para su tipo")
	ErrDocumentoCuotaExcedida  = errors.New("se superó el tamaño total permitido de documentos")
	ErrDocumentoNoEncontrado   = errors.New("documento inexistente o ajeno")
	ErrFirmaDescargaInvalida   = errors.New("el enlace de descarga venció o no es válido")
)

// CargaDocumentoRepository es el puerto de persistencia de la carga de documentos. Las
// operaciones con bloqueo deben ejecutarse dentro de una transacción del TxManager.
type CargaDocumentoRepository interface {
	// BloquearPostulacion hace SELECT ... FOR UPDATE y devuelve el estado. ErrNotFound si no
	// existe, ErrAccesoDenegado si es de otro postulante.
	BloquearPostulacion(ctx context.Context, postulacionID, postulanteID string) (estado string, err error)
	TotalBytes(ctx context.Context, postulacionID string) (int64, error)
	// Insertar traduce el rechazo del CHECK de tamaño a ErrDocumentoExcedeTamano.
	Insertar(ctx context.Context, d *Documento) error
	ObtenerDePostulacion(ctx context.Context, docID, postulacionID string) (*Documento, error)
	EliminarDePostulacion(ctx context.Context, docID, postulacionID string) error
	ObtenerDePostulante(ctx context.Context, docID, postulanteID string) (*Documento, error)
}

// AlmacenDocumentos guarda los archivos fuera del directorio público (SDD 9.7).
type AlmacenDocumentos interface {
	// EscribirTemporal copia el contenido a un archivo temporal y devuelve su referencia y bytes.
	EscribirTemporal(r io.Reader) (ref string, bytes int64, err error)
	// Confirmar mueve el temporal a su nombre definitivo (atómico) y devuelve la ruta guardada.
	Confirmar(ref, nombreAlmacenado string) (ruta string, err error)
	Descartar(ref string) error
	Eliminar(nombreAlmacenado string) error
	Abrir(nombreAlmacenado string) (io.ReadCloser, error)
}

// FirmadorDescarga firma y verifica las URL de descarga (HMAC con vigencia corta).
type FirmadorDescarga interface {
	Firmar(docID string, vence time.Time) string
	// Verificar devuelve ErrFirmaDescargaInvalida si la firma no coincide o venció.
	Verificar(docID string, vence int64, firma string, ahora time.Time) error
}
