package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/interfaces/http/dto"
)

const (
	LocalsPostulanteID = "postulante_id"
	LocalsCedula       = "cedula"
	LocalsEmail        = "email"
	LocalsClaims       = "claims"
)

// JWTProtected valida la cabecera Authorization: Bearer <token> e inyecta el contexto del postulante (SRP-003).
func JWTProtected(authService *application.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.Failure("Cabecera Authorization no proporcionada"))
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.Failure("Formato de autorización inválido. Formato esperado: Bearer <token>"))
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.Failure("Token de autenticación no proporcionado"))
		}

		claims, err := authService.ValidarAccessToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.Failure("Token de autenticación expirado o inválido"))
		}

		// Inyección de contexto del usuario autenticado (Criterio 3 de Aceptación)
		c.Locals(LocalsPostulanteID, claims.PostulanteID)
		c.Locals(LocalsCedula, claims.Cedula)
		c.Locals(LocalsEmail, claims.Email)
		c.Locals(LocalsClaims, claims)

		return c.Next()
	}
}

// GetPostulanteID extrae de forma tipada el postulante_id inyectado en el contexto por el middleware.
func GetPostulanteID(c *fiber.Ctx) string {
	id, _ := c.Locals(LocalsPostulanteID).(string)
	return id
}

// GetUserClaims extrae los claims del token validados en el contexto.
func GetUserClaims(c *fiber.Ctx) *application.TokenClaims {
	claims, _ := c.Locals(LocalsClaims).(*application.TokenClaims)
	return claims
}
