package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// NewRequestID asigna un UUID v4 a cada solicitud para correlación en logs y auditoría.
func NewRequestID() fiber.Handler {
	return requestid.New(requestid.Config{
		Header: "X-Request-ID",
	})
}
