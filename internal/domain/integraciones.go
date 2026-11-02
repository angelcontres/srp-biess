package domain

import (
	"context"
	"time"
)

// TipoValidacion indica el dato biométrico/documental contrastado con el Registro Civil.
type TipoValidacion string

const (
	ValidacionFechaExpedicion TipoValidacion = "FECHA_EXPEDICION"
	ValidacionCodigoDactilar  TipoValidacion = "CODIGO_DACTILAR"
)

func (t TipoValidacion) EsValido() bool {
	return t == ValidacionFechaExpedicion || t == ValidacionCodigoDactilar
}

// SectorCargo representa el sector al que postula el candidato (RN elegibilidad).
type SectorCargo string

const (
	SectorAfiliados   SectorCargo = "AFILIADOS"
	SectorJubilados   SectorCargo = "JUBILADOS"
	SectorEmpleadores SectorCargo = "EMPLEADORES"
)

func (s SectorCargo) EsValido() bool {
	return s == SectorAfiliados || s == SectorJubilados || s == SectorEmpleadores
}

// IdentityResult es la respuesta del Registro Civil (SRS 6.2.1).
type IdentityResult struct {
	Match           bool      `json:"match"`
	Cedula          string    `json:"cedula"`
	Nombres         string    `json:"nombres,omitempty"`
	Apellidos       string    `json:"apellidos,omitempty"`
	Genero          string    `json:"genero,omitempty"`
	FechaNacimiento time.Time `json:"fecha_nacimiento,omitempty"`
	Nacionalidad    string    `json:"nacionalidad,omitempty"`
	EstadoCivil     string    `json:"estado_civil,omitempty"`
}

// EsEcuatoriano aplica la regla bloqueante de nacionalidad (RN 1.9.1).
func (r *IdentityResult) EsEcuatoriano() bool {
	return r.Nacionalidad == "ECUATORIANA"
}

// EmploymentQuality es la calidad del ciudadano según el Historial Laboral del IESS.
type EmploymentQuality struct {
	Cedula      string `json:"cedula"`
	EsAfiliado  bool   `json:"es_afiliado"`
	EsJubilado  bool   `json:"es_jubilado"`
	EsEmpleador bool   `json:"es_empleador"`
}

// CumpleSector indica si la calidad IESS habilita la postulación al sector.
func (q *EmploymentQuality) CumpleSector(sector SectorCargo) bool {
	switch sector {
	case SectorAfiliados:
		return q.EsAfiliado
	case SectorJubilados:
		return q.EsJubilado
	case SectorEmpleadores:
		return q.EsEmpleador
	}
	return false
}

// EligibilityError es el rechazo reglamentario por sector (SRS 07, mensajes institucionales).
type EligibilityError struct {
	Code    string
	Sector  SectorCargo
	Mensaje string
}

func (e *EligibilityError) Error() string { return e.Mensaje }

func (e *EligibilityError) Unwrap() error { return ErrIneligibleCandidate }

func NewEligibilityError(sector SectorCargo) *EligibilityError {
	switch sector {
	case SectorAfiliados:
		return &EligibilityError{
			Code:    "POS_NO_ELEGIBLE_AFI",
			Sector:  sector,
			Mensaje: "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS, NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS.",
		}
	case SectorJubilados:
		return &EligibilityError{
			Code:    "POS_NO_ELEGIBLE_JUB",
			Sector:  sector,
			Mensaje: "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación.",
		}
	default:
		return &EligibilityError{
			Code:    "POS_NO_ELEGIBLE_EMP",
			Sector:  SectorEmpleadores,
			Mensaje: "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador.",
		}
	}
}

// IdentityProvider es el puerto hacia el Registro Civil (ADR-010).
type IdentityProvider interface {
	Validate(ctx context.Context, cedula string, tipo TipoValidacion, valor string) (*IdentityResult, error)
}

// EmploymentHistoryProvider es el puerto hacia el Historial Laboral del IESS (ADR-010).
type EmploymentHistoryProvider interface {
	GetQuality(ctx context.Context, cedula string, sector SectorCargo) (*EmploymentQuality, error)
}
