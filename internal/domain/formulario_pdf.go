package domain

import (
	"errors"
	"time"
)

// Formulario de postulación en PDF (SRS 3.5.21 RF-EF12-20, SDD 7.5.8): todas las secciones, el
// texto de consentimiento, la firma del postulante y sus datos de identificación.

// CampoPDF es una línea "Etiqueta: valor" del formulario.
type CampoPDF struct {
	Etiqueta string
	Valor    string
}

// BloquePDF agrupa campos bajo un título (una sección o un ítem de una lista repetible).
type BloquePDF struct {
	Titulo     string
	Campos     []CampoPDF
	Subbloques []BloquePDF
}

// FormularioPDF son los datos que se renderizan.
type FormularioPDF struct {
	CodigoUnico string
	FechaEnvio  time.Time
	Cargo       string
	// Estado coincide con el comprobante y "Mi postulación" (SDD 8.5.9).
	Estado              string
	Identificacion      []CampoPDF
	Secciones           []BloquePDF
	VersionTerminos     string
	TextoConsentimiento string
}

// GeneradorFormulario renderiza el formulario a bytes PDF (SDD 5: PdfGenerator, infraestructura).
type GeneradorFormulario interface {
	Generar(f FormularioPDF) ([]byte, error)
}

// ErrFormularioNoGenerado: si el PDF no se genera, el envío no se confirma (SRS 3.5.20).
var ErrFormularioNoGenerado = errors.New("no fue posible generar el formulario de postulación")

// NombreFormularioPDF es el nombre de descarga del formulario (SDD 7.5.8).
func NombreFormularioPDF(codigoUnico string) string {
	return "Postulacion-BIESS-" + codigoUnico + ".pdf"
}
