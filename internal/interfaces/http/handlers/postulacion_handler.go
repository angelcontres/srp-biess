package handlers

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/interfaces/http/dto"
	"srp-biess/internal/interfaces/http/middleware"
)

// PostulacionHandler expone la integración mínima del formulario (SRP-013, SDD 7.3.4).
type PostulacionHandler struct {
	service *application.FormularioService
}

func NewPostulacionHandler(service *application.FormularioService) *PostulacionHandler {
	return &PostulacionHandler{service: service}
}

func fallo(c *fiber.Ctx, status int, code, mensaje string) error {
	return c.Status(status).JSON(fiber.Map{"success": false, "code": code, "error": mensaje})
}

// responderError traduce los errores de dominio a códigos del catálogo SDD 16.6.
func responderError(c *fiber.Ctx, err error) error {
	var valErr *domain.ErrorValidacionSeccion
	if errors.As(err, &valErr) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success":  false,
			"code":     "VALIDATION_ERROR",
			"error":    "Los datos de la sección no son válidos.",
			"detalles": valErr.Detalles,
		})
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return fallo(c, fiber.StatusNotFound, "POS_NO_ENCONTRADA", "No existe una postulación para este postulante.")
	case errors.Is(err, domain.ErrPostulacionExists):
		return fallo(c, fiber.StatusConflict, "POS_YA_EXISTE", "Ya existe una postulación para esta convocatoria.")
	case errors.Is(err, domain.ErrFaseCerrada):
		return fallo(c, fiber.StatusConflict, "POS_FASE_CERRADA", "La fase de postulación no está habilitada.")
	case errors.Is(err, domain.ErrConflictoVersion):
		return fallo(c, fiber.StatusConflict, "POS_CONFLICTO_VERSION", "La postulación fue modificada en otra sesión. Recargue la página.")
	case errors.Is(err, domain.ErrPostulacionCerrada):
		return fallo(c, fiber.StatusConflict, "POS_CERRADA", "La postulación ya no admite cambios.")
	case errors.Is(err, domain.ErrSeccionInvalida):
		return fallo(c, fiber.StatusUnprocessableEntity, "VALIDATION_ERROR", "Los datos de la sección no son válidos.")
	case errors.Is(err, domain.ErrDocumentoTipoNoCargable):
		return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Este tipo de documento no se puede cargar.")
	case errors.Is(err, domain.ErrDocumentoNoPDF):
		return fallo(c, fiber.StatusUnsupportedMediaType, "DOC_TIPO_INVALIDO", "El archivo no es un PDF válido. Seleccione un archivo en formato PDF.")
	case errors.Is(err, domain.ErrDocumentoExcedeTamano):
		return fallo(c, fiber.StatusRequestEntityTooLarge, "DOC_TAMANO_EXCEDIDO", "El archivo supera el tamaño máximo permitido para este documento.")
	case errors.Is(err, domain.ErrDocumentoCuotaExcedida):
		return fallo(c, fiber.StatusRequestEntityTooLarge, "DOC_TOTAL_EXCEDIDO", "Sus documentos superan el tamaño total permitido para la postulación. Elimine o reemplace algún documento.")
	case errors.Is(err, domain.ErrDocumentoNoEncontrado):
		return fallo(c, fiber.StatusNotFound, "DOC_NO_ENCONTRADO", "El documento no existe.")
	case errors.Is(err, domain.ErrFirmaDescargaInvalida):
		return fallo(c, fiber.StatusForbidden, "DOC_ENLACE_INVALIDO", "El enlace de descarga venció o no es válido. Vuelva a abrir el documento.")
	case errors.Is(err, domain.ErrAccesoDenegado):
		return fallo(c, fiber.StatusForbidden, "POS_ACCESO_DENEGADO", "No tiene permiso para modificar esta postulación.")
	case errors.Is(err, domain.ErrCargoInvalido):
		return fallo(c, fiber.StatusBadRequest, "POS_CARGO_INVALIDO", "El cargo seleccionado no está disponible.")
	case errors.Is(err, domain.ErrConfirmacionCargo):
		return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Debe confirmar la selección del cargo.")
	case errors.Is(err, domain.ErrCargoYaConfirmado):
		return fallo(c, fiber.StatusConflict, "POS_CARGO_CONFIRMADO", "El cargo de esta postulación ya fue confirmado.")
	case errors.Is(err, domain.ErrExternalUnavailable):
		return fallo(c, fiber.StatusServiceUnavailable, "INTEGRACION_NO_DISPONIBLE", "No es posible validar su calidad laboral en este momento.")
	case errors.Is(err, application.ErrPersistenciaNoDisponible):
		return fallo(c, fiber.StatusServiceUnavailable, "SERVICIO_NO_DISPONIBLE", "El servicio de postulaciones no está disponible.")
	default:
		return fallo(c, fiber.StatusInternalServerError, "ERROR_INTERNO", "Ocurrió un problema en el sistema.")
	}
}

func postulanteID(c *fiber.Ctx) string {
	id, _ := c.Locals(middleware.LocalsPostulanteID).(string)
	return id
}

// ObtenerActual implementa GET /postulaciones/actual: postulación vigente o 404.
func (h *PostulacionHandler) ObtenerActual(c *fiber.Ctx) error {
	estado, err := h.service.ObtenerActual(c.Context(), postulanteID(c))
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(dto.Success(estado))
}

// Crear implementa POST /postulaciones: borrador único por convocatoria (RF-EF12-01).
func (h *PostulacionHandler) Crear(c *fiber.Ctx) error {
	estado, err := h.service.CrearBorrador(c.Context(), postulanteID(c))
	if err != nil {
		return responderError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.Success(estado, "Borrador creado"))
}

// ObtenerDetalle implementa GET /postulaciones/:id: estado, datos verificados y secciones guardadas.
func (h *PostulacionHandler) ObtenerDetalle(c *fiber.Ctx) error {
	detalle, err := h.service.ObtenerDetalle(c.Context(), c.Params("id"), postulanteID(c))
	if err != nil {
		return responderError(c, err)
	}
	secciones := make(map[string]json.RawMessage, len(detalle.Secciones))
	for nombre, datos := range detalle.Secciones {
		secciones[nombre] = json.RawMessage(datos)
	}
	return c.JSON(dto.Success(fiber.Map{
		"postulacion": detalle.Postulacion,
		"verificados": detalle.Verificados,
		"secciones":   secciones,
	}))
}

// GuardarSeccion devuelve el handler de guardado de una sección. El cuerpo es el objeto de la
// sección más `version` (control optimista) y `completa`, que se retiran antes de persistir.
func (h *PostulacionHandler) GuardarSeccion(seccion string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var cuerpo map[string]json.RawMessage
		if err := json.Unmarshal(c.Body(), &cuerpo); err != nil || cuerpo == nil {
			return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de petición inválido.")
		}
		var version int
		if err := json.Unmarshal(cuerpo["version"], &version); err != nil || version < 1 {
			return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Falta la versión de la postulación.")
		}
		var completa bool
		if raw, ok := cuerpo["completa"]; ok {
			_ = json.Unmarshal(raw, &completa)
		}
		delete(cuerpo, "version")
		delete(cuerpo, "completa")

		datos, err := json.Marshal(cuerpo)
		if err != nil {
			return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de petición inválido.")
		}
		estado, err := h.service.GuardarSeccion(c.Context(), c.Params("id"), postulanteID(c), seccion, datos, completa, version)
		if err != nil {
			return responderError(c, err)
		}
		return c.JSON(dto.Success(estado, "Progreso guardado"))
	}
}
