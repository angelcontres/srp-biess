package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"srp-biess/internal/domain"
)

// Guardar postulación (SDD 7.5.6, SDD 10.11): BORRADOR → LISTA_PARA_ENVIO con el consentimiento.
// Es el paso previo obligatorio de /enviar, que solo acepta LISTA_PARA_ENVIO.

// ErrPostulacionIncompleta: faltan secciones, la elegibilidad o el consentimiento (422 POS_INCOMPLETA).
var ErrPostulacionIncompleta = errors.New("la postulación no está completa")

// PostulacionIncompletaError detalla qué falta, en texto para el postulante.
type PostulacionIncompletaError struct {
	Faltantes []string
}

func (e *PostulacionIncompletaError) Error() string {
	return ErrPostulacionIncompleta.Error() + ": " + strings.Join(e.Faltantes, "; ")
}

func (e *PostulacionIncompletaError) Unwrap() error { return ErrPostulacionIncompleta }

// MaxVersionTerminos es el largo de consentimiento.version_terminos (varchar(20)).
const MaxVersionTerminos = 20

// ErrVersionTerminosInvalida: versión vacía o más larga que la columna (400, nunca 500).
var ErrVersionTerminosInvalida = errors.New("la versión de los términos no es válida")

// ErrConsentimientosInvalidos: demasiados textos o un texto demasiado largo (400).
var ErrConsentimientosInvalidos = errors.New("los textos de consentimiento no son válidos")

func versionTerminosValida(v string) bool {
	v = strings.TrimSpace(v)
	return v != "" && len([]rune(v)) <= MaxVersionTerminos
}

// seccionesOpcionales no bloquean el envío: Capacitación puede quedar sin registros (igual que la SPA).
var seccionesOpcionales = map[string]bool{"capacitacion": true}

type GuardarParaEnvioRequest struct {
	PostulacionID        string
	PostulanteID         string
	AceptaConsentimiento bool
	VersionTerminos      string
	// TextosConsentimiento son los textos que el postulante aceptó; el PDF imprime estos mismos.
	TextosConsentimiento []string
	Solicitud            domain.ContextoSolicitud
}

// Máximo de consentimientos y de caracteres por texto: evita guardar cuerpos arbitrarios.
const (
	maxConsentimientos     = 10
	maxTextoConsentimiento = 2000
)

type GuardarParaEnvioResult struct {
	Estado           string `json:"estado"`
	PorcentajeAvance int    `json:"porcentajeAvance"`
}

// GuardarParaEnvio valida la completitud con los estados de sección del servidor y deja la
// postulación LISTA_PARA_ENVIO. Repetirlo sobre una postulación ya lista es idempotente.
//
// La completitud de cada sección la informa hoy el guardado por sección; la validación de reglas
// en el servidor llega con SRP-017d (#41).
func (s *PostulacionService) GuardarParaEnvio(ctx context.Context, req GuardarParaEnvioRequest) (*GuardarParaEnvioResult, error) {
	if req.PostulanteID == "" {
		return nil, domain.ErrUnauthorized
	}
	if s.lector == nil {
		return nil, ErrPersistenciaNoDisponible
	}
	if req.AceptaConsentimiento && strings.TrimSpace(req.VersionTerminos) != "" && !versionTerminosValida(req.VersionTerminos) {
		return nil, ErrVersionTerminosInvalida
	}
	textos := make([]string, 0, len(req.TextosConsentimiento))
	for _, t := range req.TextosConsentimiento {
		if t = strings.TrimSpace(t); t != "" {
			textos = append(textos, t)
		}
	}
	if len(textos) > maxConsentimientos {
		return nil, ErrConsentimientosInvalidos
	}
	for _, t := range textos {
		if len([]rune(t)) > maxTextoConsentimiento {
			return nil, ErrConsentimientosInvalidos
		}
	}
	var result GuardarParaEnvioResult
	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		p, err := s.postulacionRepo.GetByIDForUpdate(txCtx, req.PostulacionID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return ErrPostulacionNoEncontrada
			}
			return fmt.Errorf("error al consultar postulación: %w", err)
		}
		if p.PostulanteID != req.PostulanteID {
			return domain.ErrAccesoDenegado
		}
		switch p.Estado {
		case domain.EstadoListaParaEnvio:
			result = GuardarParaEnvioResult{Estado: string(domain.EstadoListaParaEnvio), PorcentajeAvance: 100}
			return nil
		case domain.EstadoEnviada:
			return ErrPostulacionYaEnviada
		case domain.EstadoBorrador:
		default:
			return domain.ErrPostulacionCerrada
		}

		detalle, err := s.lector.ObtenerDetalle(txCtx, p.ID, p.PostulanteID)
		if err != nil {
			return fmt.Errorf("error leyendo la postulación: %w", err)
		}
		var faltantes []string
		if e := detalle.Postulacion.Elegibilidad; e == nil || !e.Elegible {
			faltantes = append(faltantes, "Confirme un cargo para el que cumpla los requisitos de elegibilidad.")
		}
		for _, seccion := range domain.SeccionesFormulario {
			if !seccionesOpcionales[seccion] && detalle.Postulacion.Secciones[seccion] != domain.SeccionCompleta {
				faltantes = append(faltantes, "Complete la sección "+titulosSeccion[seccion]+".")
			}
		}
		if !req.AceptaConsentimiento || strings.TrimSpace(req.VersionTerminos) == "" {
			faltantes = append(faltantes, "Acepte los consentimientos vigentes.")
		}
		if len(faltantes) > 0 {
			return &PostulacionIncompletaError{Faltantes: faltantes}
		}

		if err := s.postulacionRepo.SaveConsentimiento(txCtx, p.ID, strings.TrimSpace(req.VersionTerminos)); err != nil {
			return fmt.Errorf("error al registrar consentimiento: %w", err)
		}
		if len(textos) > 0 {
			if err := s.postulacionRepo.SaveTextosConsentimiento(txCtx, p.ID, textos); err != nil {
				return fmt.Errorf("error al registrar los textos de consentimiento: %w", err)
			}
		}
		if err := s.postulacionRepo.UpdateEstado(txCtx, p.ID, domain.EstadoListaParaEnvio); err != nil {
			return fmt.Errorf("error al actualizar el estado: %w", err)
		}
		if err := s.auditarEstado(txCtx, req.Solicitud, domain.AccionPostulacionLista, p,
			map[string]any{"estado": string(domain.EstadoListaParaEnvio)}); err != nil {
			return fmt.Errorf("error auditando el cambio de estado: %w", err)
		}
		result = GuardarParaEnvioResult{Estado: string(domain.EstadoListaParaEnvio), PorcentajeAvance: 100}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}
