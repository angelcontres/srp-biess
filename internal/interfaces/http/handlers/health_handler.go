package handlers

import (
	"github.com/gofiber/fiber/v2"
	"srp-biess/internal/interfaces/http/dto"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	return c.JSON(dto.Success(fiber.Map{
		"service": "srp-biess-api",
		"version": "1.0.0",
		"status":  "healthy",
	}, "Servicio operativo"))
}
