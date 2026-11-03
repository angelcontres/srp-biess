package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/external"
	interfacesHTTP "srp-biess/internal/interfaces/http"
	"srp-biess/internal/interfaces/http/handlers"
)

const urlValidarIdentidad = "/api/v1/auth/registro/validar-identidad"

// spyIdentityProvider delega en el mock y cuenta las llamadas al Registro Civil.
type spyIdentityProvider struct {
	inner *external.MockIdentityProvider
	calls int
}

func (s *spyIdentityProvider) Validate(ctx context.Context, cedula string, tipo domain.TipoValidacion, valor string) (*domain.IdentityResult, error) {
	s.calls++
	return s.inner.Validate(ctx, cedula, tipo, valor)
}

func setupRegistroApp() (*fiber.App, *application.AuthService, *spyIdentityProvider) {
	spy := &spyIdentityProvider{inner: external.NewMockIdentityProvider()}
	authService := application.NewAuthService(nil, nil).WithIdentityProvider(spy)
	app := fiber.New()
	interfacesHTTP.SetupRoutes(app, interfacesHTTP.RouterConfig{
		AuthHandler: handlers.NewAuthHandler(authService),
		AuthService: authService,
	})
	return app, authService, spy
}

func postValidarIdentidad(t *testing.T, app *fiber.App, cedula, tipo, valor string) (*http.Response, map[string]interface{}) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"cedula": cedula, "tipoValidacion": tipo, "valor": valor})
	req := httptest.NewRequest(http.MethodPost, urlValidarIdentidad, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var jsonRes map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&jsonRes)
	return resp, jsonRes
}

func TestAuthHandler_ValidarIdentidadRegistro(t *testing.T) {
	t.Run("200 sin Authorization, con datos y tokenValidacion de 10 min", func(t *testing.T) {
		app, svc, _ := setupRegistroApp()
		resp, jsonRes := postValidarIdentidad(t, app, "0926687856", "FECHA_EXPEDICION", "2019-05-14")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d; want 200 (%v)", resp.StatusCode, jsonRes)
		}
		data := jsonRes["data"].(map[string]interface{})
		datos := data["datos"].(map[string]interface{})
		if data["identidadValidada"] != true || datos["nombres"] != "MARIA JOSE" || datos["fechaNacimiento"] != "1985-03-21" {
			t.Errorf("respuesta inesperada: %v", data)
		}
		claims, err := svc.ValidarTokenValidacion(data["tokenValidacion"].(string))
		if err != nil {
			t.Fatalf("tokenValidacion inválido: %v", err)
		}
		if claims.Cedula != "0926687856" || claims.ExpiresAt.Sub(claims.IssuedAt.Time) != 10*time.Minute {
			t.Errorf("claims inesperados: cedula=%q ttl=%v", claims.Cedula, claims.ExpiresAt.Sub(claims.IssuedAt.Time))
		}
	})

	t.Run("422 genérico: mismo mensaje para dato erróneo y cédula desconocida", func(t *testing.T) {
		app, _, _ := setupRegistroApp()
		resp, mismatch := postValidarIdentidad(t, app, "0926687856", "CODIGO_DACTILAR", "X0000X0000")
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d; want 422", resp.StatusCode)
		}
		_, desconocida := postValidarIdentidad(t, app, "1710034065", "FECHA_EXPEDICION", "1999-01-01")
		if mismatch["error"] != desconocida["error"] || mismatch["data"] != nil {
			t.Errorf("el 422 debe ser idéntico y sin datos: %v vs %v", mismatch, desconocida)
		}
	})

	t.Run("422 fail-fast por Módulo 10 sin llamar al Registro Civil (SRP-008)", func(t *testing.T) {
		app, _, spy := setupRegistroApp()
		resp, _ := postValidarIdentidad(t, app, "0926687857", "FECHA_EXPEDICION", "2019-05-14")
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d; want 422", resp.StatusCode)
		}
		if spy.calls != 0 {
			t.Errorf("llamadas al Registro Civil = %d; want 0", spy.calls)
		}
	})

	t.Run("503 cuando el Registro Civil no está disponible", func(t *testing.T) {
		app, _, _ := setupRegistroApp()
		resp, _ := postValidarIdentidad(t, app, external.CedulaServicioCaido, "FECHA_EXPEDICION", "2019-05-14")
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("status = %d; want 503", resp.StatusCode)
		}
	})
}

func TestAuthHandler_RegistroBloqueadoDevuelve423YRetryAfter(t *testing.T) {
	app, svc, _ := setupRegistroApp()
	store := &blockedAttemptStore{}
	svc.WithAuthAttemptStore(store)

	resp, body := postValidarIdentidad(t, app, "0926687856", "FECHA_EXPEDICION", "2019-05-14")
	if resp.StatusCode != http.StatusLocked {
		t.Fatalf("status = %d; want 423 (%v)", resp.StatusCode, body)
	}
	if got := resp.Header.Get("Retry-After"); got != "600" {
		t.Fatalf("Retry-After = %q; want 600", got)
	}
	if body["error"] != "registro fallido, inténtelo nuevamente en 10 minutos" {
		t.Fatalf("mensaje inesperado: %v", body["error"])
	}
}

type blockedAttemptStore struct{}

func (blockedAttemptStore) CheckBlocked(context.Context, domain.AuthFlow, string) (time.Duration, error) {
	return 10 * time.Minute, nil
}

func (blockedAttemptStore) RecordFailure(context.Context, domain.AuthFlow, string) (bool, time.Duration, error) {
	return true, 10 * time.Minute, nil
}

func (blockedAttemptStore) Reset(context.Context, domain.AuthFlow, string) error {
	return nil
}
