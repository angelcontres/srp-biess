package domain

import (
	"context"
	"errors"
)

// Confirmación del cargo y elegibilidad por sector (SRP-017a, RF-EF12-09, RF-EF12-10, SDD 7.5.4).

// Estados de elegibilidad persistidos en postulacion.elegibilidad_estado.
const (
	ElegibilidadPendiente  = "PENDIENTE"
	ElegibilidadElegible   = "ELEGIBLE"
	ElegibilidadNoElegible = "NO_ELEGIBLE"
)

var (
	ErrCargoInvalido     = errors.New("el cargo seleccionado no está disponible")
	ErrCargoYaConfirmado = errors.New("el cargo de la postulación ya fue confirmado")
	ErrConfirmacionCargo = errors.New("la selección del cargo requiere confirmación")
	// ErrAccesoDenegado: la postulación existe pero pertenece a otro postulante (403, auditado).
	ErrAccesoDenegado = errors.New("la postulación no pertenece al postulante autenticado")
)

// ContextoCargo reúne lo necesario para validar la confirmación antes de consultar al IESS.
type ContextoCargo struct {
	Estado             string
	Version            int
	ElegibilidadEstado string
	Cedula             string
	Sector             SectorCargo
}

// ResultadoElegibilidad es lo que se persiste al confirmar el cargo.
type ResultadoElegibilidad struct {
	CargoID  int
	Elegible bool
	Motivo   string
}

// CargoRepository es el puerto de persistencia de la confirmación del cargo.
// Todas las operaciones filtran por postulanteID: un postulante solo accede a su postulación.
type CargoRepository interface {
	// ContextoConfirmacion devuelve ErrNotFound si la postulación no existe, ErrAccesoDenegado
	// si es de otro postulante y ErrCargoInvalido si el cargo no existe o no está activo.
	ContextoConfirmacion(ctx context.Context, id, postulanteID string, cargoID int) (*ContextoCargo, error)
	// ConfirmarCargo guarda cargo y elegibilidad con control de versión; si no es elegible,
	// la postulación pasa a NO_ELEGIBLE en la misma sentencia. Se une a la transacción del contexto.
	ConfirmarCargo(ctx context.Context, id, postulanteID string, r ResultadoElegibilidad, version int) error
	// EstadoPostulacion devuelve el estado que consume la SPA.
	EstadoPostulacion(ctx context.Context, id, postulanteID string) (*EstadoFormulario, error)
}

// MensajeElegible es el texto que se muestra cuando el postulante cumple la calidad del sector
// (SDD 7.5.4 para afiliados; SRS 3.5.11 para jubilados y empleadores).
func MensajeElegible(sector SectorCargo) string {
	switch sector {
	case SectorAfiliados:
		return "Postulante afiliado al IESS"
	case SectorJubilados:
		return "Postulante jubilado"
	default:
		return "Postulante empleador"
	}
}
