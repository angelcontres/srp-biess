package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/interfaces/http/middleware"
)

func TestJWTProtectedMiddleware(t *testing.T) {
	authService := application.NewAuthService(nil, nil)

	app := fiber.New()
	app.Get("/api/v1/protegido/perfil", middleware.JWTProtected(authService), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"postulante_id": middleware.GetPostulanteID(c),
			"cedula":        c.Locals("cedula"),
			"email":         c.Locals("email"),
		})
	})

	t.Run("Petición sin cabecera Authorization retorna 401 estructurado", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/protegido/perfil", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error ejecutando request: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("Petición con esquema no-Bearer retorna 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/protegido/perfil", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error ejecutando request: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("Petición con token manipulado retorna 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/protegido/perfil", nil)
		req.Header.Set("Authorization", "Bearer token_falso_invalido")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error ejecutando request: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("Petición con Access Token válido inyecta contexto correctamente (200 OK)", func(t *testing.T) {
		postulanteID := "postulante-uuid-777"
		cedula := "0926687856"
		email := "aspirante@biess.gob.ec"

		token, err := authService.GenerarAccessToken(postulanteID, cedula, email)
		if err != nil {
			t.Fatalf("error generando token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/protegido/perfil", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error ejecutando request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusOK)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("error decodificando respuesta: %v", err)
		}

		if body["postulante_id"] != postulanteID {
			t.Errorf("postulante_id = %v; want %s", body["postulante_id"], postulanteID)
		}
		if body["cedula"] != cedula {
			t.Errorf("cedula = %v; want %s", body["cedula"], cedula)
		}
		if body["email"] != email {
			t.Errorf("email = %v; want %s", body["email"], email)
		}
	})
}
