package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"srp-biess/internal/interfaces/http/dto"
)

// NewRecover previene caídas del proceso capturando panics y retornando HTTP 500 JSON.
func NewRecover() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			log.Printf("[PANIC RECOVERED] %v\n", e)
			_ = c.Status(fiber.StatusInternalServerError).JSON(dto.Failure("Error interno inesperado en el servidor"))
		},
	})
}
