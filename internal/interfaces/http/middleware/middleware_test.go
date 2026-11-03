package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/interfaces/http/middleware"
)

func TestMiddlewares(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.NewRequestID())
	app.Use(middleware.NewCORS())
	app.Use(middleware.NewStructuredLogger())
	app.Use(middleware.NewRecover())

	app.Get("/test-ok", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Get("/test-panic", func(c *fiber.Ctx) error {
		panic("simulated critical crash")
	})

	t.Run("Headers CORS y Request-ID presentes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-ok", nil)
		req.Header.Set("Origin", "http://localhost:5173") // Simulación de petición desde Frontend SPA Vite

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error en request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusOK)
		}
		if resp.Header.Get("X-Request-ID") == "" {
			t.Errorf("se esperaba cabecera X-Request-ID generada")
		}
		if resp.Header.Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
			t.Errorf("se esperaba Access-Control-Allow-Origin: http://localhost:5173")
		}
	})

	t.Run("Recover captura panic y responde 500 JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-panic", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error en request: %v", err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusInternalServerError)
		}

		body, _ := io.ReadAll(resp.Body)
		if len(body) == 0 {
			t.Errorf("se esperaba cuerpo JSON en la respuesta de error")
		}
	})
}
