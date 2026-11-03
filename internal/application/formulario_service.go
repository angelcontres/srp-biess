package application

import (
	"context"
	"encoding/json"
	"errors"

	"srp-biess/internal/domain"
)

// ErrPersistenciaNoDisponible se devuelve cuando la API arrancó sin PostgreSQL.
var ErrPersistenciaNoDisponible = errors.New("la persistencia de postulaciones no está disponible")

// FormularioService orquesta la creación, consulta y guardado por secciones del
// formulario de postulación (SRP-013, integración mínima de SDD 7.3.4).
type FormularioService struct {
	repo domain.FormularioRepository
}

func NewFormularioService(repo domain.FormularioRepository) *FormularioService {
	return &FormularioService{repo: repo}
}

func (s *FormularioService) disponible() error {
	if s.repo == nil {
		return ErrPersistenciaNoDisponible
	}
	return nil
}

// ObtenerActual devuelve la postulación vigente del postulante o domain.ErrNotFound.
func (s *FormularioService) ObtenerActual(ctx context.Context, postulanteID string) (*domain.EstadoFormulario, error) {
	if err := s.disponible(); err != nil {
		return nil, err
	}
	return s.repo.ObtenerActual(ctx, postulanteID)
}

// CrearBorrador crea el borrador único por postulante y convocatoria (RF-EF12-01).
func (s *FormularioService) CrearBorrador(ctx context.Context, postulanteID string) (*domain.EstadoFormulario, error) {
	if err := s.disponible(); err != nil {
		return nil, err
	}
	if _, err := s.repo.ObtenerActual(ctx, postulanteID); err == nil {
		return nil, domain.ErrPostulacionExists
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	return s.repo.CrearBorrador(ctx, postulanteID)
}

// ObtenerDetalle devuelve estado, datos verificados y valores de cada sección.
func (s *FormularioService) ObtenerDetalle(ctx context.Context, id, postulanteID string) (*domain.DetalleFormulario, error) {
	if err := s.disponible(); err != nil {
		return nil, err
	}
	return s.repo.ObtenerDetalle(ctx, id, postulanteID)
}

// GuardarSeccion persiste una sección con control de versión optimista (RF-EF12-02).
//
// TODO(backend: SRP-005): la completitud la informa el cliente; RF-EF12-24 exige que el
// servidor valide las reglas de cada sección antes de marcarla como COMPLETA.
func (s *FormularioService) GuardarSeccion(
	ctx context.Context, id, postulanteID, seccion string, datos []byte, completa bool, version int,
) (*domain.EstadoFormulario, error) {
	if err := s.disponible(); err != nil {
		return nil, err
	}
	if !domain.EsSeccionFormulario(seccion) {
		return nil, domain.ErrSeccionInvalida
	}
	var objeto map[string]json.RawMessage
	if err := json.Unmarshal(datos, &objeto); err != nil {
		return nil, domain.ErrSeccionInvalida
	}
	return s.repo.GuardarSeccion(ctx, id, postulanteID, seccion, datos, completa, version)
}
