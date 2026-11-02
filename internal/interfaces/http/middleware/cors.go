package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// NewCORS inicializa la política CORS para permitir la interacción segura con la SPA React.
func NewCORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, http://127.0.0.1:300",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
	})
}
