package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/interfaces/http/handlers"
	"srp-biess/internal/interfaces/http/middleware"
)

func TestAuthHandler(t *testing.T) {
	authService := application.NewAuthService(nil, nil)
	authHandler := handlers.NewAuthHandler(authService)

	app := fiber.New()
	app.Post("/api/v1/auth/validate-cedula", authHandler.ValidarCedula)
	app.Get("/api/v1/auth/me", middleware.JWTProtected(authService), authHandler.Me)

	t.Run("Validar cédula ecuatoriana exitosa", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cedula": "0926687856",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate-cedula", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("Validar cédula inválida retorna 422", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cedula": "0926687850",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate-cedula", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("Endpoint /me protegido con JWT", func(t *testing.T) {
		token, err := authService.GenerarAccessToken("user-456", "0926687856", "test@biess.gob.ec")
		if err != nil {
			t.Fatalf("error generando token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusOK)
		}
	})
}
