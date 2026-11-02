package handlers

import (
	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/interfaces/http/dto"
)

type AuthHandler struct {
	authService *application.AuthService
}

func NewAuthHandler(authService *application.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type ValidarCedulaRequest struct {
	Cedula string `json:"cedula"`
}

func (h *AuthHandler) ValidarCedula(c *fiber.Ctx) error {
	var req ValidarCedulaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Failure("Cuerpo de petición inválido"))
	}

	if err := h.authService.ValidarIdentidad(req.Cedula); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(dto.Failure(err.Error()))
	}

	return c.JSON(dto.Success(fiber.Map{
		"cedula": req.Cedula,
		"valida": true,
	}, "Cédula verificada correctamente con algoritmo Módulo 10"))
}

// Me retorna la identidad del postulante autenticado mediante el contexto inyectado (SRP-003).
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	postulanteID, _ := c.Locals("postulante_id").(string)
	cedula, _ := c.Locals("cedula").(string)
	email, _ := c.Locals("email").(string)

	return c.JSON(dto.Success(fiber.Map{
		"postulante_id": postulanteID,
		"cedula":        cedula,
		"email":         email,
	}, "Sesión activa y válida"))
}
