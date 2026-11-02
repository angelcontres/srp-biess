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

// SetupRoutes registra los endpoints de la API desacoplados por versión (SDD 07).
func SetupRoutes(app *fiber.App, cfg RouterConfig) {
	// Endpoint raíz y health check
	app.Get("/health", cfg.HealthHandler.Check)

	// Grupo API v1
	v1 := app.Group("/api/v1")

	// -----------------------------------------------------------------
	// Autenticación & Sesión (ADR-006, SDD 07, SRP-003)
	// -----------------------------------------------------------------
	if cfg.AuthHandler != nil {
		auth := v1.Group("/auth")

		// Endpoints públicos
		auth.Post("/validate-cedula", cfg.AuthHandler.ValidarCedula)
		auth.Post("/login", cfg.AuthHandler.Login)
		auth.Post("/refresh", cfg.AuthHandler.Refresh)
		// TODO(backend): falta auth.Post("/registro/validar-identidad", ...) (SDD 7.4.1), que el
		// frontend de SRP-012 ya consume. Detalle y alcance en el TODO de cmd/api/main.go.

		// Endpoints protegidos con Access Token JWT (15 min)
		if cfg.AuthService != nil {
			protected := auth.Group("", middleware.JWTProtected(cfg.AuthService))
			protected.Get("/me", cfg.AuthHandler.Me)
			protected.Post("/logout", cfg.AuthHandler.Logout)
		}
	}

	// -----------------------------------------------------------------
	// Elegibilidad IESS (EF-11, EF-12, SRP-009)
	// -----------------------------------------------------------------
	if cfg.EligibilityHandler != nil {
		elegibilidad := v1.Group("/elegibilidad")
		elegibilidad.Post("/verificar", cfg.EligibilityHandler.Verificar)
	}
}
