package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"srp-biess/internal/domain"
)

// DefaultIESSTimeout es la espera máxima del Historial Laboral (IESS_TIMEOUT, SDD 16).
const DefaultIESSTimeout = 5 * time.Second

// CargoService confirma el cargo de la postulación y registra la elegibilidad (SRP-017a, SDD 10.7).
type CargoService struct {
	tx           domain.TxManager
	repo         domain.CargoRepository
	auditoria    domain.AuditoriaRepository
	elegibilidad *EligibilityService
	timeout      time.Duration
}

func NewCargoService(
	tx domain.TxManager, repo domain.CargoRepository, auditoria domain.AuditoriaRepository,
	elegibilidad *EligibilityService, timeout time.Duration,
) *CargoService {
	if timeout <= 0 {
		timeout = DefaultIESSTimeout
	}
	return &CargoService{tx: tx, repo: repo, auditoria: auditoria, elegibilidad: elegibilidad, timeout: timeout}
}

// ConfirmarCargo valida la postulación, consulta al IESS y guarda el resultado con su auditoría
// en una sola transacción (RF-EF12-09, RF-EF12-10, SDD 10.7).
//
//   - Elegible: devuelve el estado actualizado y error nil.
//   - No elegible: la postulación queda NO_ELEGIBLE; devuelve el estado y un *domain.EligibilityError
//     con el mensaje institucional (422 en HTTP, SDD 7.5.4).
//   - Proveedor caído, timeout o error: domain.ErrExternalUnavailable sin escribir nada, para
//     que el postulante pueda reintentar.
func (s *CargoService) ConfirmarCargo(
	ctx context.Context, sol domain.ContextoSolicitud, id, postulanteID string, cargoID int, confirmacion bool, version int,
) (*domain.EstadoFormulario, error) {
	if s.repo == nil || s.tx == nil || s.auditoria == nil {
		return nil, ErrPersistenciaNoDisponible
	}
	if !confirmacion {
		return nil, domain.ErrConfirmacionCargo
	}
	if cargoID <= 0 {
		return nil, domain.ErrCargoInvalido
	}

	c, err := s.repo.ContextoConfirmacion(ctx, id, postulanteID, cargoID)
	if errors.Is(err, domain.ErrAccesoDenegado) {
		s.auditarAccesoDenegado(ctx, sol, id, postulanteID)
	}
	if err != nil {
		return nil, err
	}
	// Se rechaza antes de consultar al IESS; la sentencia de guardado vuelve a comprobarlo.
	switch {
	case c.Estado != "BORRADOR":
		return nil, domain.ErrPostulacionCerrada
	case c.ElegibilidadEstado != domain.ElegibilidadPendiente:
		return nil, domain.ErrCargoYaConfirmado
	case c.Version != version:
		return nil, domain.ErrConflictoVersion
	}

	consulta, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	res := domain.ResultadoElegibilidad{CargoID: cargoID}
	_, err = s.elegibilidad.Validate(consulta, c.Cedula, c.Sector)
	var noElegible *domain.EligibilityError
	switch {
	case err == nil:
		res.Elegible = true
		res.Motivo = domain.MensajeElegible(c.Sector)
	case errors.As(err, &noElegible):
		res.Motivo = noElegible.Mensaje
	default:
		// Caída, timeout o error del proveedor: no se marca NO_ELEGIBLE (RF-EF12-10).
		return nil, fmt.Errorf("%w: %v", domain.ErrExternalUnavailable, err)
	}

	err = s.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.ConfirmarCargo(txCtx, id, postulanteID, res, version); err != nil {
			return err
		}
		return s.auditoria.Registrar(txCtx, eventoCargo(sol, id, postulanteID, c, res, version))
	})
	if err != nil {
		return nil, err
	}

	estado, err := s.repo.EstadoPostulacion(ctx, id, postulanteID)
	if err != nil {
		return nil, err
	}
	if noElegible != nil {
		return estado, noElegible
	}
	return estado, nil
}

// eventoCargo registra antes y después sin datos personales: la cédula no se audita.
func eventoCargo(
	sol domain.ContextoSolicitud, id, postulanteID string, c *domain.ContextoCargo, res domain.ResultadoElegibilidad, version int,
) domain.EventoAuditoria {
	accion, estado, elegibilidad := domain.AccionCargoConfirmado, c.Estado, domain.ElegibilidadElegible
	if !res.Elegible {
		accion, estado, elegibilidad = domain.AccionPostulacionNoElegible, "NO_ELEGIBLE", domain.ElegibilidadNoElegible
	}
	return domain.EventoAuditoria{
		ActorID: postulanteID, ActorTipo: domain.ActorPostulante,
		Accion: accion, Entidad: "postulacion", EntidadID: id,
		Antes: map[string]any{
			"cargoId": nil, "estado": c.Estado, "elegibilidadEstado": c.ElegibilidadEstado, "version": version,
		},
		Despues: map[string]any{
			"cargoId": res.CargoID, "sector": c.Sector, "estado": estado,
			"elegibilidadEstado": elegibilidad, "motivo": res.Motivo, "version": version + 1,
		},
		Solicitud: sol,
	}
}

// El acceso denegado se audita fuera de la transacción (SDD 9.9); si falla, solo se registra en el log.
func (s *CargoService) auditarAccesoDenegado(ctx context.Context, sol domain.ContextoSolicitud, id, postulanteID string) {
	err := s.auditoria.Registrar(ctx, domain.EventoAuditoria{
		ActorID: postulanteID, ActorTipo: domain.ActorPostulante,
		Accion: domain.AccionAccesoDenegado, Entidad: "postulacion", EntidadID: id,
		Despues:   map[string]any{"operacion": "PUT /postulaciones/{id}/cargo"},
		Solicitud: sol,
	})
	if err != nil {
		log.Printf("[WARN] no se pudo auditar el acceso denegado a la postulación %s: %v", id, err)
	}
}
