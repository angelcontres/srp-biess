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
		// TODO(backend: SRP-008): esta ruta no existe en el SDD. SDD 7.3.4 / 7.6.1 la define como
		// POST /api/v1/validaciones/cedula (pública, fuera de /auth) y responde
		// { valida, provincia, codigoProvincia, tipo, algoritmo }; el 422 usa VAL_CEDULA_INVALIDA.
		// El frontend (SRP-012) no la consume todavía.
		auth.Post("/validate-cedula", cfg.AuthHandler.ValidarCedula)
		auth.Post("/login", cfg.AuthHandler.Login)
		auth.Post("/refresh", cfg.AuthHandler.Refresh)
		// Se registra antes del grupo protegido para que no la alcance el middleware JWT (ver TODO SRP-003).
		auth.Post("/registro/validar-identidad", cfg.AuthHandler.ValidarIdentidadRegistro)

		// Endpoints protegidos con Access Token JWT (15 min)
		if cfg.AuthService != nil {
			// TODO(backend: SRP-003): Group("", JWTProtected) aplica el middleware a TODO /api/v1/auth/*,
			// así que cualquier ruta inexistente bajo /auth responde 401 "Cabecera Authorization no
			// proporcionada" en vez de 404 (le pasó al registro de SRP-012 antes de existir /registro/validar-identidad).
			// Sugerencia: registrar el middleware por ruta, p. ej.
			//   auth.Get("/me", middleware.JWTProtected(cfg.AuthService), cfg.AuthHandler.Me)
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
