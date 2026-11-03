package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"srp-biess/internal/domain"
)

// ErrPersistenciaNoDisponible se devuelve cuando la API arrancó sin PostgreSQL.
var ErrPersistenciaNoDisponible = errors.New("la persistencia de postulaciones no está disponible")

// FormularioService orquesta la creación, consulta y guardado por secciones del
// formulario de postulación (SRP-013, integración mínima de SDD 7.3.4).
type FormularioService struct {
	repo         domain.FormularioRepository
	catalogoRepo domain.CatalogoRepository
}

func NewFormularioService(repo domain.FormularioRepository) *FormularioService {
	return &FormularioService{repo: repo}
}

// WithCatalogoRepository inyecta el repositorio de catálogos para resolver dinámicamente identificadores oficiales.
func (s *FormularioService) WithCatalogoRepository(catRepo domain.CatalogoRepository) *FormularioService {
	s.catalogoRepo = catRepo
	return s
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
// Valida las reglas de completitud en el servidor si completa == true (RF-EF12-24, SRP-017d).
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

	if completa {
		var ctxPostulante domain.ContextoPostulante
		detalle, err := s.repo.ObtenerDetalle(ctx, id, postulanteID)
		if err != nil {
			return nil, err
		}
		if detalle != nil && detalle.Postulacion != nil && detalle.Postulacion.Estado != "" && detalle.Postulacion.Estado != "BORRADOR" {
			return nil, domain.ErrPostulacionCerrada
		}
		if detalle != nil && detalle.Verificados != nil {
			ctxPostulante.CedulaPostulante = detalle.Verificados.Cedula
			ctxPostulante.CorreoPrincipal = detalle.Verificados.Correo
		}

		// Identificación de Ecuador por catálogo o valor estándar de la semilla (id 52 en db/init.sql / ISO "EC")
		ctxPostulante.PaisEcuadorID = "52"
		if s.catalogoRepo != nil {
			if paises, err := s.catalogoRepo.GetPaises(ctx); err == nil {
				for _, p := range paises {
					if strings.EqualFold(p.Codigo, "EC") {
						ctxPostulante.PaisEcuadorID = p.ID
						break
					}
				}
			}
		}

		if err := domain.ValidarSeccion(seccion, datos, ctxPostulante); err != nil {
			return nil, err
		}
	}

	return s.repo.GuardarSeccion(ctx, id, postulanteID, seccion, datos, completa, version)
}
