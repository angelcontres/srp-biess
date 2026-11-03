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
	t.Setenv("APP_ENV", "development")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")

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
		if resp.Header.Get("Access-Control-Allow-Credentials") != "true" {
			t.Errorf("se esperaba Access-Control-Allow-Credentials: true")
		}
		if resp.Header.Get("Access-Control-Expose-Headers") != "Retry-After,X-Request-ID" {
			t.Errorf("se esperaba exposición de Retry-After y X-Request-ID")
		}
	})

	t.Run("Origen no permitido no recibe cabecera CORS", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-ok", nil)
		req.Header.Set("Origin", "https://malicioso.example")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error en request: %v", err)
		}
		if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("no se esperaba Access-Control-Allow-Origin, se recibió %q", got)
		}
	})

	t.Run("La lista configurada por entorno cambia la política", func(t *testing.T) {
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")

		configuredApp := fiber.New()
		configuredApp.Use(middleware.NewCORS())
		configuredApp.Get("/test-ok", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test-ok", nil)
		req.Header.Set("Origin", "https://portal.example")
		resp, err := configuredApp.Test(req)
		if err != nil {
			t.Fatalf("error en request: %v", err)
		}
		if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "https://portal.example" {
			t.Errorf("origen permitido = %q", got)
		}
	})

	t.Run("Preflight expone métodos y cabeceras esperados", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test-ok", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		req.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error en request: %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusNoContent)
		}
		if got := resp.Header.Get("Access-Control-Allow-Methods"); got != "GET,POST,PUT,PATCH,DELETE,OPTIONS" {
			t.Errorf("métodos permitidos = %q", got)
		}
		if got := resp.Header.Get("Access-Control-Allow-Headers"); got != "Origin,Content-Type,Accept,Authorization,X-Request-ID" {
			t.Errorf("cabeceras permitidas = %q", got)
		}
	})

	t.Run("Producción sin configuración rechaza todos los orígenes", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("CORS_ALLOWED_ORIGINS", "")

		productionApp := fiber.New()
		productionApp.Use(middleware.NewCORS())
		productionApp.Get("/test-ok", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test-ok", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		resp, err := productionApp.Test(req)
		if err != nil {
			t.Fatalf("error en request: %v", err)
		}
		if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("no se esperaba Access-Control-Allow-Origin, se recibió %q", got)
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
