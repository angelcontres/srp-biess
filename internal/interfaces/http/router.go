package http

import (
	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/interfaces/http/handlers"
	"srp-biess/internal/interfaces/http/middleware"
)

// RouterConfig consolida los handlers y servicios inyectados para el registro de rutas.
type RouterConfig struct {
	HealthHandler      *handlers.HealthHandler
	AuthHandler        *handlers.AuthHandler
	EligibilityHandler *handlers.EligibilityHandler
	AuthService        *application.AuthService
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

	// Ruta protegida con JWT (SRP-003)
	if cfg.AuthService != nil && cfg.AuthHandler != nil {
		auth.Get("/me", middleware.JWTProtected(cfg.AuthService), cfg.AuthHandler.Me)
	}

	// Elegibilidad IESS (EF-11, EF-12)
	if cfg.EligibilityHandler != nil {
		elegibilidad := v1.Group("/elegibilidad")
		elegibilidad.Post("/verificar", cfg.EligibilityHandler.Verificar)
	}
}
