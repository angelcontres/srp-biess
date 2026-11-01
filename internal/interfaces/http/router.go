package http

import (
	"github.com/gofiber/fiber/v2"
	"srp-biess/internal/interfaces/http/handlers"
)

type RouterConfig struct {
	HealthHandler *handlers.HealthHandler
	AuthHandler   *handlers.AuthHandler
}

// SetupRoutes registra los endpoints de la API desacoplados por versión.
func SetupRoutes(app *fiber.App, cfg RouterConfig) {
	// Endpoint raíz y health check
	app.Get("/health", cfg.HealthHandler.Check)

	// Grupo API v1
	v1 := app.Group("/api/v1")

	// Auth & Registro Civil Mock
	auth := v1.Group("/auth")
	auth.Post("/validate-cedula", cfg.AuthHandler.ValidarCedula)
}
