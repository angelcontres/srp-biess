package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	interfacesHTTP "srp-biess/internal/interfaces/http"
	"srp-biess/internal/interfaces/http/handlers"
)

func setupTestApp() (*fiber.App, *application.AuthService) {
	authService := application.NewAuthService(nil, nil)
	authHandler := handlers.NewAuthHandler(authService)

	app := fiber.New()
	interfacesHTTP.SetupRoutes(app, interfacesHTTP.RouterConfig{
		AuthHandler: authHandler,
		AuthService: authService,
	})

	return app, authService
}

func TestAuthHandler_Lifecycle(t *testing.T) {
	app, _ := setupTestApp()

	cedulaValida := "0926687856"
	passwordCorrecta := "BiessPostulante2026*"
	passwordIncorrecta := "ClaveErronea123"

	var accessToken string
	var refreshCookie *http.Cookie

	t.Run("1. Login con contraseña errónea retorna 401 institucional", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cedula":   cedulaValida,
			"password": passwordIncorrecta,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("2. Login exitoso emite Access Token (15 min) y Cookie HttpOnly (8 h)", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"cedula":   cedulaValida,
			"password": passwordCorrecta,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d; want %d", resp.StatusCode, http.StatusOK)
		}

		// Validar cuerpo JSON con Access Token
		var jsonRes map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&jsonRes)
		data := jsonRes["data"].(map[string]interface{})
		accessToken = data["access_token"].(string)

		if accessToken == "" {
			t.Errorf("access_token no debe estar vacío")
		}
		if data["expires_in"].(float64) != 900 {
			t.Errorf("expires_in = %v; want 900", data["expires_in"])
		}

		// Validar Cookie HttpOnly
		cookies := resp.Cookies()
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshCookie = c
				break
			}
		}

		if refreshCookie == nil {
			t.Fatalf("se esperaba cookie refresh_token en la respuesta")
		}
		if !refreshCookie.HttpOnly {
			t.Errorf("la cookie refresh_token debe ser HttpOnly")
		}
		if refreshCookie.Path != "/api/v1/auth" {
			t.Errorf("path esperado /api/v1/auth, obtenido: %s", refreshCookie.Path)
		}
	})

	t.Run("3. Acceso a ruta protegida /me con Access Token válido", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("4. Renovación /refresh con Cookie HttpOnly y rotación de token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.AddCookie(refreshCookie)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d; want %d", resp.StatusCode, http.StatusOK)
		}

		var jsonRes map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&jsonRes)
		data := jsonRes["data"].(map[string]interface{})
		newAccessToken := data["access_token"].(string)

		if newAccessToken == "" || newAccessToken == accessToken {
			t.Errorf("se esperaba un nuevo access token generado")
		}

		// Verificar que se emitió una nueva cookie (Rotación de Refresh Token)
		var newCookie *http.Cookie
		for _, c := range resp.Cookies() {
			if c.Name == "refresh_token" {
				newCookie = c
				break
			}
		}
		if newCookie == nil || newCookie.Value == refreshCookie.Value {
			t.Errorf("se esperaba rotación de la cookie refresh_token con nuevo valor")
		}
		refreshCookie = newCookie
	})

	t.Run("5. Renovación sin cookie retorna 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("6. Logout revoca token y elimina cookie HttpOnly", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.AddCookie(refreshCookie)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d; want %d", resp.StatusCode, http.StatusOK)
		}

		// Validar que la cookie fue eliminada (MaxAge negativo o expirada)
		cleared := false
		for _, c := range resp.Cookies() {
			if c.Name == "refresh_token" && (c.MaxAge < 0 || c.Value == "") {
				cleared = true
				break
			}
		}
		if !cleared {
			t.Errorf("se esperaba cookie refresh_token invalidada en respuesta de logout")
		}
	})

	t.Run("7. Intentar renovar con token revocado falla con 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.AddCookie(refreshCookie)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d; want 401 tras revocar el token", resp.StatusCode)
		}
	})

	t.Run("8. Validar cédula ecuatoriana endpoint", func(t *testing.T) {
		body := strings.NewReader(`{"cedula":"0926687856"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate-cedula", body)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want 200", resp.StatusCode)
		}
	})
}
