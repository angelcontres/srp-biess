package domain

import (
	"context"
	"errors"
	"time"
)

// Integración mínima del formulario de postulación (SRP-013, SDD 7.3.4).
// Cada sección se persiste como un documento JSON en `postulacion_seccion` hasta que
// se normalice en las tablas de SDD 4.6 (pendiente de SRP-005).

// SeccionesFormulario son las secciones editables, en el orden de SRS 3.5.1 (sin Resumen).
var SeccionesFormulario = []string{
	"personal", "domicilio", "familiares", "cargo", "formacion",
	"experiencia", "publicaciones", "capacitacion", "inhabilidades",
}

// EsSeccionFormulario indica si el nombre corresponde a una sección editable.
func EsSeccionFormulario(seccion string) bool {
	for _, s := range SeccionesFormulario {
		if s == seccion {
			return true
		}
	}
	return false
}

// EstadoSeccion: COMPLETA, INCOMPLETA o SIN_INICIAR (el stepper se guía por el servidor, SDD 8.5.4).
type EstadoSeccion string

const (
	SeccionCompleta   EstadoSeccion = "COMPLETA"
	SeccionIncompleta EstadoSeccion = "INCOMPLETA"
	SeccionSinIniciar EstadoSeccion = "SIN_INICIAR"
)

// Elegibilidad es el resultado de la validación del cargo (RF-EF12-10).
type Elegibilidad struct {
	Elegible bool      `json:"elegible"`
	CargoID  int       `json:"cargoId"`
	Mensaje  string    `json:"mensaje"`
	Fecha    time.Time `json:"fecha"`
}

// EstadoFormulario es el estado de la postulación que consume la SPA.
type EstadoFormulario struct {
	ID             string                   `json:"id"`
	Estado         string                   `json:"estado"`
	Version        int                      `json:"version"`
	Secciones      map[string]EstadoSeccion `json:"secciones"`
	UltimoGuardado *time.Time               `json:"ultimoGuardado"`
	Elegibilidad   *Elegibilidad            `json:"elegibilidad"`
	// Código único y fecha de envío (SDD 8.5.10): vacíos mientras no se envíe.
	CodigoUnico        string     `json:"codigoUnico,omitempty"`
	FechaEnvio         *time.Time `json:"fechaEnvio"`
	SeccionReanudacion string     `json:"seccionReanudacion"`
}

// DatosVerificados son los datos de identidad del postulante, de solo lectura (RF-EF12-03).
type DatosVerificados struct {
	Cedula          string `json:"cedula"`
	Nombres         string `json:"nombres"`
	Apellidos       string `json:"apellidos"`
	Genero          string `json:"genero"`
	FechaNacimiento string `json:"fechaNacimiento"`
	Nacionalidad    string `json:"nacionalidad"`
	Correo          string `json:"correo"`
}

// DetalleFormulario agrupa el estado, los datos verificados y los valores guardados por sección.
type DetalleFormulario struct {
	Postulacion *EstadoFormulario `json:"postulacion"`
	Verificados *DatosVerificados `json:"verificados"`
	// Secciones contiene el JSON guardado de cada sección (bytes crudos, sin dependencia de encoding).
	Secciones map[string][]byte `json:"-"`
}

var (
	ErrConflictoVersion   = errors.New("la postulación fue modificada en otra sesión")
	ErrPostulacionCerrada = errors.New("la postulación ya no admite cambios")
	ErrFaseCerrada        = errors.New("la fase de postulación no está habilitada")
	ErrSeccionInvalida    = errors.New("sección del formulario no válida")
)

// FormularioRepository es el puerto de persistencia del formulario de postulación.
// Todas las operaciones filtran por postulanteID: un postulante solo accede a su postulación.
type FormularioRepository interface {
	ObtenerActual(ctx context.Context, postulanteID string) (*EstadoFormulario, error)
	CrearBorrador(ctx context.Context, postulanteID string) (*EstadoFormulario, error)
	ObtenerDetalle(ctx context.Context, id, postulanteID string) (*DetalleFormulario, error)
	GuardarSeccion(ctx context.Context, id, postulanteID, seccion string, datos []byte, completa bool, version int) (*EstadoFormulario, error)
}

// EventoEstado es una entrada del historial de estados de la postulación (SDD 8.5.10).
type EventoEstado struct {
	Estado string    `json:"estado"`
	Fecha  time.Time `json:"fecha"`
}
