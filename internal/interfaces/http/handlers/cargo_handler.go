package handlers

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/interfaces/http/dto"
)

// CargoHandler expone PUT /postulaciones/:id/cargo (SRP-017a, SDD 7.5.4).
type CargoHandler struct {
	service *application.CargoService
}

func NewCargoHandler(service *application.CargoService) *CargoHandler {
	return &CargoHandler{service: service}
}

type confirmarCargoRequest struct {
	CargoID      int  `json:"cargoId"`
	Confirmacion bool `json:"confirmacion"`
	Version      int  `json:"version"`
}

// Confirmar implementa SDD 7.5.4 y 10.7:
//
//   - 200 elegible: { elegible, mensaje, requiereRespaldo, version } del SDD, más { postulacion,
//     elegibilidad }, que consume la SPA (frontend/src/features/postulacion/services/httpGateway.ts).
//   - 422 no elegible: { code: POS_NO_ELEGIBLE_*, detail, estado: NO_ELEGIBLE } del SDD, más el
//     mismo `data` para que la SPA actualice la postulación sin otra consulta.
func (h *CargoHandler) Confirmar(c *fiber.Ctx) error {
	if h.service == nil {
		return responderError(c, application.ErrPersistenciaNoDisponible)
	}
	var req confirmarCargoRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de petición inválido.")
	}
	if req.Version < 1 {
		return fallo(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Falta la versión de la postulación.")
	}
	estado, err := h.service.ConfirmarCargo(
		c.Context(), contextoSolicitud(c), c.Params("id"), postulanteID(c), req.CargoID, req.Confirmacion, req.Version,
	)
	var noElegible *domain.EligibilityError
	if errors.As(err, &noElegible) && estado != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"code":    noElegible.Code,
			"error":   noElegible.Mensaje,
			"detail":  noElegible.Mensaje,
			"estado":  estado.Estado,
			"data":    fiber.Map{"postulacion": estado, "elegibilidad": estado.Elegibilidad},
		})
	}
	if err != nil {
		return responderError(c, err)
	}
	cuerpo := fiber.Map{
		"postulacion":      estado,
		"elegibilidad":     estado.Elegibilidad,
		"elegible":         true,
		"requiereRespaldo": true, // RF-EF12-11: la sección de respaldo sigue a la elegibilidad aprobada
		"version":          estado.Version,
	}
	if estado.Elegibilidad != nil {
		cuerpo["mensaje"] = estado.Elegibilidad.Mensaje
	}
	return c.JSON(dto.Success(cuerpo, "Cargo confirmado"))
}

// contextoSolicitud reúne IP, navegador y trace_id (X-Request-ID) para la auditoría (SDD 9.9).
func contextoSolicitud(c *fiber.Ctx) domain.ContextoSolicitud {
	return domain.ContextoSolicitud{
		IP:        c.IP(),
		UserAgent: c.Get(fiber.HeaderUserAgent),
		TraceID:   c.GetRespHeader("X-Request-ID"),
	}
}
