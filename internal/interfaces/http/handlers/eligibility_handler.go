package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/interfaces/http/dto"
)

// EligibilityHandler maneja la verificación de requisitos laborales IESS (EF-11, EF-12).
type EligibilityHandler struct {
	service *application.EligibilityService
}

func NewEligibilityHandler(service *application.EligibilityService) *EligibilityHandler {
	return &EligibilityHandler{service: service}
}

type VerificarElegibilidadRequest struct {
	Cedula string `json:"cedula"`
	Sector string `json:"sector"`
}

// Verificar valida si un postulante cumple la calidad laboral para el sector seleccionado.
func (h *EligibilityHandler) Verificar(c *fiber.Ctx) error {
	var req VerificarElegibilidadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Failure("Cuerpo de petición inválido"))
	}

	if req.Cedula == "" || req.Sector == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Failure("Los campos 'cedula' y 'sector' son obligatorios"))
	}

	sector := domain.SectorCargo(req.Sector)
	calidad, err := h.service.Validate(c.Context(), req.Cedula, sector)
	if err != nil {
		var eligErr *domain.EligibilityError
		if errors.As(err, &eligErr) {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"code":    eligErr.Code,
				"sector":  eligErr.Sector,
				"error":   eligErr.Mensaje,
				"calidad": calidad,
			})
		}

		if errors.Is(err, domain.ErrInvalidSector) {
			return c.Status(fiber.StatusBadRequest).JSON(dto.Failure("Sector de postulación inválido"))
		}

		if errors.Is(err, domain.ErrExternalUnavailable) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(dto.Failure("El servicio del IESS no se encuentra disponible temporalmente"))
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.Failure("Error al consultar el historial laboral del IESS"))
	}

	return c.JSON(dto.Success(calidad, "Candidato habilitado: cumple con la calidad requerida para el sector"))
}
