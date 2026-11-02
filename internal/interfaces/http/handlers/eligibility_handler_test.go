package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/infrastructure/external"
	"srp-biess/internal/interfaces/http/handlers"
)

func TestEligibilityHandler(t *testing.T) {
	iessMock := external.NewMockEmploymentProvider()
	eligibilityService := application.NewEligibilityService(iessMock)
	handler := handlers.NewEligibilityHandler(eligibilityService)

	app := fiber.New()
	app.Post("/api/v1/elegibilidad/verificar", handler.Verificar)

	t.Run("Postulante afiliado habilitado para sector afiliados", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cedula": "0926687856",
			"sector": "AFILIADOS",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/elegibilidad/verificar", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error en request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("Postulante no cumple calidad para sector jubilados (422 con código institucional)", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cedula": "0926687856",
			"sector": "JUBILADOS",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/elegibilidad/verificar", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error en request: %v", err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("Sector inválido retorna 400 Bad Request", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cedula": "0926687856",
			"sector": "INVALIDO",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/elegibilidad/verificar", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error en request: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("Servicio IESS caído retorna 503", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cedula": external.CedulaServicioCaido,
			"sector": "AFILIADOS",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/elegibilidad/verificar", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error en request: %v", err)
		}
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusServiceUnavailable)
		}
	})
}
