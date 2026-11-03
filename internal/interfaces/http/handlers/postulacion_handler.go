package handlers

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/interfaces/http/dto"
	"srp-biess/internal/interfaces/http/middleware"
)

// PostulacionHandler expone la integración mínima del formulario (SRP-013, SDD 7.3.4).
type PostulacionHandler struct {
	service *application.FormularioService
	submit  *application.PostulacionService
}

func NewPostulacionHandler(service *application.FormularioService, submit ...*application.PostulacionService) *PostulacionHandler {
	handler := &PostulacionHandler{service: service}
	if len(submit) > 0 {
		handler.submit = submit[0]
	}
	return handler
}

func fallo(c *fiber.Ctx, status int, code, mensaje string) error {
	return c.Status(status).JSON(fiber.Map{"success": false, "code": code, "error": mensaje})
}

// responderError traduce los errores de dominio a códigos del catálogo SDD 16.6.
func responderError(c *fiber.Ctx, err error) error {
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
	case errors.Is(err, application.ErrConsentimientosInvalidos):
		return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Los textos de consentimiento enviados no son válidos.")
	case errors.Is(err, application.ErrVersionTerminosInvalida):
		return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "La versión de los términos aceptados no es válida.")
	case errors.Is(err, application.ErrPostulacionNoEncontrada):
		return fallo(c, fiber.StatusNotFound, "POS_NO_ENCONTRADA", "No existe una postulación para este postulante.")
	case errors.Is(err, domain.ErrFormularioNoGenerado):
		return fallo(c, fiber.StatusServiceUnavailable, "POS_PDF_NO_GENERADO", "No fue posible generar su formulario. Intente nuevamente.")
	case errors.Is(err, domain.ErrUnauthorized):
		return fallo(c, fiber.StatusUnauthorized, "AUTH_TOKEN_EXPIRADO", "Su sesión no es válida. Ingrese nuevamente.")
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
	case errors.Is(err, application.ErrIdempotencyKeyRequerida):
		return fallo(c, fiber.StatusBadRequest, "IDEMPOTENCY_KEY_REQUERIDA", "La cabecera Idempotency-Key es obligatoria.")
	case errors.Is(err, application.ErrIdempotencyKeyConflicto):
		return fallo(c, fiber.StatusConflict, "IDEMPOTENCY_KEY_CONFLICTO", "La cabecera Idempotency-Key ya fue utilizada para otra postulación.")
	case errors.Is(err, application.ErrConsentimientoRequerido):
		return fallo(c, fiber.StatusBadRequest, "CONSENTIMIENTO_REQUERIDO", "Debe aceptar el consentimiento vigente para enviar la postulación.")
	case errors.Is(err, application.ErrPostulacionYaEnviada):
		return fallo(c, fiber.StatusConflict, "POS_YA_ENVIADA", "La postulación ya fue enviada.")
	case errors.Is(err, application.ErrPostulacionInvalida):
		return fallo(c, fiber.StatusConflict, "POS_NO_LISTA_PARA_ENVIO", "La postulación no está lista para ser enviada.")
	case errors.Is(err, domain.ErrExternalUnavailable):
		return fallo(c, fiber.StatusServiceUnavailable, "INTEGRACION_NO_DISPONIBLE", "No es posible validar su calidad laboral en este momento.")
	case errors.Is(err, application.ErrPersistenciaNoDisponible):
		return fallo(c, fiber.StatusServiceUnavailable, "SERVICIO_NO_DISPONIBLE", "El servicio de postulaciones no está disponible.")
	default:
		return fallo(c, fiber.StatusInternalServerError, "ERROR_INTERNO", "Ocurrió un problema en el sistema.")
	}
}

// Enviar implementa POST /postulaciones/:id/enviar con idempotencia transaccional.
func (h *PostulacionHandler) Enviar(c *fiber.Ctx) error {
	if h.submit == nil {
		return responderError(c, application.ErrPersistenciaNoDisponible)
	}
	key := strings.TrimSpace(c.Get("Idempotency-Key"))
	var req struct {
		AceptaConsentimiento bool   `json:"aceptaConsentimiento"`
		VersionTerminos      string `json:"versionTerminos"`
		Version              int    `json:"version"`
		// Un campo `email` en el cuerpo se ignora: la notificación va al correo registrado.
	}
	if len(c.Body()) > 0 {
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de petición inválido.")
		}
	}
	if key == "" || !req.AceptaConsentimiento || strings.TrimSpace(req.VersionTerminos) == "" {
		if key == "" {
			return responderError(c, application.ErrIdempotencyKeyRequerida)
		}
		return responderError(c, application.ErrConsentimientoRequerido)
	}
	// La postulación se vincula al postulante del JWT; el servicio rechaza una ajena (403).
	result, err := h.submit.EnviarPostulacion(c.Context(), application.EnviarPostulacionRequest{
		PostulacionID:        c.Params("id"),
		PostulanteID:         postulanteID(c),
		AceptaConsentimiento: req.AceptaConsentimiento,
		VersionTerminos:      strings.TrimSpace(req.VersionTerminos),
		IdempotencyKey:       key,
		Solicitud:            contextoSolicitud(c),
	})
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(dto.Success(fiber.Map{
		"codigoUnico": result.CodigoUnico,
		"estado":      result.Estado,
		"fechaEnvio":  result.FechaEnvio,
		"pdfUrl":      result.PDFURL,
		"documentoId": result.DocumentoID,
	}))
}

// Guardar implementa POST /postulaciones/:id/guardar (SDD 7.5.6): BORRADOR → LISTA_PARA_ENVIO.
// Si falta algo responde 422 POS_INCOMPLETA con la lista `faltantes`.
func (h *PostulacionHandler) Guardar(c *fiber.Ctx) error {
	if h.submit == nil {
		return responderError(c, application.ErrPersistenciaNoDisponible)
	}
	var req struct {
		AceptaConsentimiento bool     `json:"aceptaConsentimiento"`
		VersionTerminos      string   `json:"versionTerminos"`
		Consentimientos      []string `json:"consentimientos"`
	}
	if len(c.Body()) > 0 {
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de petición inválido.")
		}
	}
	result, err := h.submit.GuardarParaEnvio(c.Context(), application.GuardarParaEnvioRequest{
		PostulacionID:        c.Params("id"),
		PostulanteID:         postulanteID(c),
		AceptaConsentimiento: req.AceptaConsentimiento,
		VersionTerminos:      req.VersionTerminos,
		TextosConsentimiento: req.Consentimientos,
		Solicitud:            contextoSolicitud(c),
	})
	var incompleta *application.PostulacionIncompletaError
	if errors.As(err, &incompleta) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success":   false,
			"code":      "POS_INCOMPLETA",
			"error":     "Su postulación no está completa.",
			"faltantes": incompleta.Faltantes,
		})
	}
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(dto.Success(result, "Postulación lista para envío"))
}

// Historial implementa GET /postulaciones/:id/historial (SDD 8.5.10): estados con fecha y hora,
// en orden cronológico inverso.
func (h *PostulacionHandler) Historial(c *fiber.Ctx) error {
	if h.submit == nil {
		return responderError(c, application.ErrPersistenciaNoDisponible)
	}
	eventos, err := h.submit.Historial(c.Context(), c.Params("id"), postulanteID(c))
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(dto.Success(eventos))
}

// DescargarFormulario implementa GET /postulaciones/:id/pdf (SDD 7.5.8): URL estable, exige el JWT
// del dueño y entrega el formulario generado al enviar.
func (h *PostulacionHandler) DescargarFormulario(c *fiber.Ctx) error {
	if h.submit == nil {
		return responderError(c, application.ErrPersistenciaNoDisponible)
	}
	doc, archivo, err := h.submit.DescargarFormulario(c.Context(), c.Params("id"), postulanteID(c))
	if err != nil {
		return responderError(c, err)
	}
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+doc.NombreOriginal+`"`)
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.SendStream(archivo, int(doc.TamanoBytes))
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
