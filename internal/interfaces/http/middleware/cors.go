package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

const defaultDevelopmentOrigins = "http://localhost:3000, http://127.0.0.1:3000, http://localhost:5173, http://127.0.0.1:5173"

// NewCORS inicializa la política CORS para permitir la interacción segura con la SPA React.
func NewCORS() fiber.Handler {
	config := cors.Config{
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		ExposeHeaders:    "Retry-After, X-Request-ID",
		AllowCredentials: true,
	}

	origins := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if origins == "" && strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		config.AllowOriginsFunc = func(string) bool {
			return false
		}
	} else {
		if origins == "" {
			origins = defaultDevelopmentOrigins
		}
		config.AllowOrigins = origins
	}

	return cors.New(config)
}
