package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/interfaces/http/dto"
)

// CatalogoHandler expone los catálogos geográficos públicos de SDD 7.3.2 (integración de SRP-013/SRP-014).
//
// La caché cache-aside en Redis (SRP-007) se inyecta en cmd/api/main.go como decorador del repositorio.
type CatalogoHandler struct {
	service *application.CatalogoService
}

// NewCatalogoHandler recibe nil cuando la API arrancó sin PostgreSQL: responde 503.
func NewCatalogoHandler(service *application.CatalogoService) *CatalogoHandler {
	return &CatalogoHandler{service: service}
}

type itemCatalogo struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	// Codigo es el ISO 3166-1 alfa-2 del país: la SPA reconoce a Ecuador por "EC", no por su id.
	Codigo string `json:"codigo,omitempty"`
}

func (h *CatalogoHandler) responder(c *fiber.Ctx, items []itemCatalogo, err error) error {
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Failure("No se pudieron cargar los catálogos."))
	}
	if items == nil {
		items = []itemCatalogo{} // siempre un arreglo JSON, nunca null
	}
	return c.JSON(dto.Success(items))
}

func (h *CatalogoHandler) disponible(c *fiber.Ctx) bool {
	if h.service == nil {
		_ = c.Status(fiber.StatusServiceUnavailable).JSON(dto.Failure("El servicio de catálogos no está disponible."))
		return false
	}
	return true
}

func idNumerico(c *fiber.Ctx, param string) (string, bool) {
	id := c.Params(param)
	if _, err := strconv.Atoi(id); err != nil {
		_ = c.Status(fiber.StatusBadRequest).JSON(dto.Failure("Identificador de catálogo inválido."))
		return "", false
	}
	return id, true
}

// Paises implementa GET /catalogos/paises.
func (h *CatalogoHandler) Paises(c *fiber.Ctx) error {
	if !h.disponible(c) {
		return nil
	}
	paises, err := h.service.ListarPaises(c.Context())
	items := make([]itemCatalogo, 0, len(paises))
	for _, p := range paises {
		items = append(items, itemCatalogo{ID: p.ID, Nombre: p.Nombre, Codigo: p.Codigo})
	}
	return h.responder(c, items, err)
}

// Provincias implementa GET /catalogos/paises/:paisId/provincias.
func (h *CatalogoHandler) Provincias(c *fiber.Ctx) error {
	if !h.disponible(c) {
		return nil
	}
	paisID, ok := idNumerico(c, "paisId")
	if !ok {
		return nil
	}
	provincias, err := h.service.ListarProvincias(c.Context(), paisID)
	items := make([]itemCatalogo, 0, len(provincias))
	for _, p := range provincias {
		items = append(items, itemCatalogo{ID: p.ID, Nombre: p.Nombre})
	}
	return h.responder(c, items, err)
}

// Cantones implementa GET /catalogos/provincias/:provinciaId/cantones: solo los de esa provincia.
func (h *CatalogoHandler) Cantones(c *fiber.Ctx) error {
	if !h.disponible(c) {
		return nil
	}
	provinciaID, ok := idNumerico(c, "provinciaId")
	if !ok {
		return nil
	}
	cantones, err := h.service.ListarCantones(c.Context(), provinciaID)
	items := make([]itemCatalogo, 0, len(cantones))
	for _, ct := range cantones {
		items = append(items, itemCatalogo{ID: ct.ID, Nombre: ct.Nombre})
	}
	return h.responder(c, items, err)
}
