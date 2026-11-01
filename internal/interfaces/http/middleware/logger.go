package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// NewStructuredLogger genera registros HTTP estructurados en JSON (SRP-001 / SDD 03).
// Permite trazabilidad centralizada compatible con Datadog, ELK y CloudWatch.
func NewStructuredLogger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     `{"time":"${time}","trace_id":"${locals:requestid}","status":${status},"latency":"${latency}","method":"${method}","path":"${path}","ip":"${ip}","bytes_sent":${bytesSent},"error":"${error}"}` + "\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "America/Guayaquil",
	})
}
