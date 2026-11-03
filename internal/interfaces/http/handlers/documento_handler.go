package handlers

import (
	"net/url"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/interfaces/http/dto"
)

// DocumentoHandler expone la carga, el borrado y la descarga de documentos (SRP-017b, SDD 7.5.5).
//
// El corte defensivo del cuerpo lo aplica Fiber con BodyLimit (equivalente a http.MaxBytesReader
// en fasthttp); el servicio vuelve a limitar el flujo por tipo con io.LimitReader.
type DocumentoHandler struct {
	service *application.DocumentoService
}

func NewDocumentoHandler(service *application.DocumentoService) *DocumentoHandler {
	return &DocumentoHandler{service: service}
}

type documentoDTO struct {
	ID             string    `json:"id"`
	Tipo           string    `json:"tipo"`
	NombreOriginal string    `json:"nombreOriginal"`
	TamanoBytes    int64     `json:"tamanoBytes"`
	HashSHA256     string    `json:"hashSha256"`
	FechaCarga     time.Time `json:"fechaCarga"`
	URLDescarga    string    `json:"urlDescarga"`
}

func nuevoDocumentoDTO(d *application.DocumentoCargado) documentoDTO {
	return documentoDTO{
		ID: d.Documento.ID, Tipo: string(d.Documento.Tipo), NombreOriginal: d.Documento.NombreOriginal,
		TamanoBytes: d.Documento.TamanoBytes, HashSHA256: d.Documento.HashSHA256,
		FechaCarga: d.Documento.CreatedAt, URLDescarga: d.URLDescarga,
	}
}

// Subir implementa POST /postulaciones/:id/documentos (multipart: archivo, tipo).
func (h *DocumentoHandler) Subir(c *fiber.Ctx) error {
	if h.service == nil {
		return responderError(c, application.ErrPersistenciaNoDisponible)
	}
	archivo, err := c.FormFile("archivo")
	if err != nil {
		return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Debe adjuntar un archivo PDF.")
	}
	contenido, err := archivo.Open()
	if err != nil {
		return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "No se pudo leer el archivo adjunto.")
	}
	defer contenido.Close()

	cargado, err := h.service.Subir(c.Context(), contextoSolicitud(c), c.Params("id"), postulanteID(c),
		domain.TipoDocumento(c.FormValue("tipo")), archivo.Filename, contenido)
	if err != nil {
		return responderError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.Success(nuevoDocumentoDTO(cargado), "Documento cargado"))
}

// Eliminar implementa DELETE /postulaciones/:id/documentos/:docId.
func (h *DocumentoHandler) Eliminar(c *fiber.Ctx) error {
	if h.service == nil {
		return responderError(c, application.ErrPersistenciaNoDisponible)
	}
	if err := h.service.Eliminar(c.Context(), contextoSolicitud(c), c.Params("id"), postulanteID(c), c.Params("docId")); err != nil {
		return responderError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Metadatos implementa GET /postulaciones/:id/documentos/:docId: datos y una URL firmada nueva.
func (h *DocumentoHandler) Metadatos(c *fiber.Ctx) error {
	if h.service == nil {
		return responderError(c, application.ErrPersistenciaNoDisponible)
	}
	cargado, err := h.service.Metadatos(c.Context(), c.Params("id"), postulanteID(c), c.Params("docId"))
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(dto.Success(nuevoDocumentoDTO(cargado)))
}

// Descargar implementa GET /postulaciones/:id/documentos/:docId/download?exp&sig. Exige JWT
// (401) y firma vigente (403): doble validación.
func (h *DocumentoHandler) Descargar(c *fiber.Ctx) error {
	if h.service == nil {
		return responderError(c, application.ErrPersistenciaNoDisponible)
	}
	vence, err := strconv.ParseInt(c.Query("exp"), 10, 64)
	if err != nil {
		return responderError(c, domain.ErrFirmaDescargaInvalida)
	}
	doc, archivo, err := h.service.Descargar(c.Context(), c.Params("id"), postulanteID(c), c.Params("docId"), vence, c.Query("sig"))
	if err != nil {
		return responderError(c, err)
	}
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="documento.pdf"; filename*=UTF-8''`+url.PathEscape(doc.NombreOriginal))
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	c.Set(fiber.HeaderCacheControl, "no-store")
	// SendStream cierra el archivo al terminar de enviarlo.
	return c.SendStream(archivo, int(doc.TamanoBytes))
}
