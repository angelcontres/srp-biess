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
	PostulacionHandler *handlers.PostulacionHandler
	CargoHandler       *handlers.CargoHandler
	CatalogoHandler    *handlers.CatalogoHandler
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
		// Compatibilidad con el endpoint anterior; el contrato canónico está fuera de /auth.
		auth.Post("/validate-cedula", func(c *fiber.Ctx) error {
			c.Set("Deprecation", "true")
			return c.Next()
		}, cfg.AuthHandler.ValidarCedula)
		auth.Post("/login", cfg.AuthHandler.Login)
		auth.Post("/refresh", cfg.AuthHandler.Refresh)
		auth.Post("/registro/validar-identidad", cfg.AuthHandler.ValidarIdentidadRegistro)

		// Endpoints protegidos con Access Token JWT (15 min)
		if cfg.AuthService != nil {
			auth.Get("/me", middleware.JWTProtected(cfg.AuthService), cfg.AuthHandler.Me)
			auth.Post("/logout", middleware.JWTProtected(cfg.AuthService), cfg.AuthHandler.Logout)
		}
	}

	if cfg.AuthHandler != nil {
		v1.Post("/validaciones/cedula", cfg.AuthHandler.ValidarCedula)
	}

	// -----------------------------------------------------------------
	// Elegibilidad IESS (EF-11, EF-12, SRP-009)
	// -----------------------------------------------------------------
	if cfg.EligibilityHandler != nil {
		elegibilidad := v1.Group("/elegibilidad")
		elegibilidad.Post("/verificar", cfg.EligibilityHandler.Verificar)
	}

	// -----------------------------------------------------------------
	// Catálogos geográficos públicos (SDD 7.3.2), integración de SRP-013/SRP-014.
	// -----------------------------------------------------------------
	if cfg.CatalogoHandler != nil {
		catalogos := v1.Group("/catalogos")
		catalogos.Get("/paises", cfg.CatalogoHandler.Paises)
		catalogos.Get("/paises/:paisId/provincias", cfg.CatalogoHandler.Provincias)
		catalogos.Get("/provincias/:provinciaId/cantones", cfg.CatalogoHandler.Cantones)
	}

	// -----------------------------------------------------------------
	// Formulario de postulación (SDD 7.3.4): integración mínima de SRP-013.
	// TODO(backend: SRP-017): documentos (#40) y envío (#43).
	// -----------------------------------------------------------------
	if cfg.PostulacionHandler != nil && cfg.AuthService != nil {
		h := cfg.PostulacionHandler
		postulaciones := v1.Group("/postulaciones", middleware.JWTProtected(cfg.AuthService))
		postulaciones.Get("/actual", h.ObtenerActual)
		postulaciones.Post("/", h.Crear)
		postulaciones.Get("/:id", h.ObtenerDetalle)

		// Guardado por sección: una sola operación expuesta en las rutas del SDD.
		postulaciones.Patch("/:id/informacion-personal", h.GuardarSeccion("personal"))
		postulaciones.Put("/:id/domicilio", h.GuardarSeccion("domicilio"))
		postulaciones.Put("/:id/familiares", h.GuardarSeccion("familiares"))
		postulaciones.Put("/:id/respaldo-organizacional", h.GuardarSeccion("cargo"))
		postulaciones.Put("/:id/formacion-academica", h.GuardarSeccion("formacion"))
		postulaciones.Put("/:id/experiencia", h.GuardarSeccion("experiencia"))
		postulaciones.Put("/:id/publicaciones", h.GuardarSeccion("publicaciones"))
		postulaciones.Put("/:id/capacitaciones", h.GuardarSeccion("capacitacion"))
		postulaciones.Put("/:id/inhabilidades", h.GuardarSeccion("inhabilidades"))

		// Confirmación del cargo con elegibilidad IESS (SRP-017a, SDD 7.5.4).
		if cfg.CargoHandler != nil {
			postulaciones.Put("/:id/cargo", cfg.CargoHandler.Confirmar)
		}
	}
}
