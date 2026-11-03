package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"srp-biess/internal/application"
	"srp-biess/internal/infrastructure/cache"
	"srp-biess/internal/infrastructure/external"
	interfacesHTTP "srp-biess/internal/interfaces/http"
	"srp-biess/internal/interfaces/http/handlers"
)

const (
	msgRegistroBloqueado = "Registro fallido, inténtelo nuevamente en 10 minutos."
	msgLoginBloqueado    = "Cuenta bloqueada temporalmente. Intente nuevamente en 10 minutos."
)

// setupBloqueoApp levanta las rutas de auth con el control de intentos sobre miniredis (SRP-006).
func setupBloqueoApp(t *testing.T) (*fiber.App, *miniredis.Miniredis, *spyIdentityProvider) {
	t.Helper()
	mr := miniredis.RunT(t)
	// Sin reintentos para que el caso "Redis caído" responda dentro del timeout de app.Test.
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1, DialTimeout: 200 * time.Millisecond})
	t.Cleanup(func() { _ = rdb.Close() })

	spy := &spyIdentityProvider{inner: external.NewMockIdentityProvider()}
	authService := application.NewAuthService(nil, nil).
		WithIdentityProvider(spy).
		WithAttemptLimiter(cache.NewRedisAttemptLimiter(rdb, "dev"))
	app := fiber.New()
	interfacesHTTP.SetupRoutes(app, interfacesHTTP.RouterConfig{
		AuthHandler: handlers.NewAuthHandler(authService),
		AuthService: authService,
	})
	return app, mr, spy
}

func postLogin(t *testing.T, app *fiber.App, cedula, password string) (*http.Response, map[string]interface{}) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"cedula": cedula, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var jsonRes map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&jsonRes)
	return resp, jsonRes
}

func esperarBloqueo(t *testing.T, resp *http.Response, jsonRes map[string]interface{}, code, mensaje string) {
	t.Helper()
	if resp.StatusCode != http.StatusLocked {
		t.Fatalf("status = %d; want 423 (%v)", resp.StatusCode, jsonRes)
	}
	if got := resp.Header.Get("Retry-After"); got != "600" {
		t.Errorf("Retry-After = %q; want 600", got)
	}
	if jsonRes["code"] != code || jsonRes["error"] != mensaje || jsonRes["retryAfterSeconds"] != float64(600) {
		t.Errorf("cuerpo inesperado: %v", jsonRes)
	}
}

func TestBloqueoRegistro(t *testing.T) {
	t.Run("al 3er dato erróneo responde 423 con el mensaje del RF y Retry-After", func(t *testing.T) {
		app, _, _ := setupBloqueoApp(t)
		for i := 1; i <= 2; i++ {
			if resp, _ := postValidarIdentidad(t, app, "0926687856", "CODIGO_DACTILAR", "X0000X0000"); resp.StatusCode != http.StatusUnprocessableEntity {
				t.Fatalf("intento %d: status = %d; want 422", i, resp.StatusCode)
			}
		}
		resp, jsonRes := postValidarIdentidad(t, app, "0926687856", "CODIGO_DACTILAR", "X0000X0000")
		esperarBloqueo(t, resp, jsonRes, "AUTH_REGISTRO_BLOQUEADO", msgRegistroBloqueado)
	})

	t.Run("durante el bloqueo rechaza incluso datos correctos sin llamar al Registro Civil", func(t *testing.T) {
		app, _, spy := setupBloqueoApp(t)
		for i := 0; i < 3; i++ {
			postValidarIdentidad(t, app, "0926687856", "CODIGO_DACTILAR", "X0000X0000")
		}
		llamadas := spy.calls
		resp, _ := postValidarIdentidad(t, app, "0926687856", "FECHA_EXPEDICION", "2019-05-14")
		if resp.StatusCode != http.StatusLocked {
			t.Fatalf("status = %d; want 423", resp.StatusCode)
		}
		if spy.calls != llamadas {
			t.Errorf("no debe consultarse el Registro Civil durante el bloqueo")
		}
	})

	t.Run("tras 10 minutos permite un nuevo intento", func(t *testing.T) {
		app, mr, _ := setupBloqueoApp(t)
		for i := 0; i < 3; i++ {
			postValidarIdentidad(t, app, "0926687856", "CODIGO_DACTILAR", "X0000X0000")
		}
		mr.FastForward(10 * time.Minute)
		if resp, _ := postValidarIdentidad(t, app, "0926687856", "FECHA_EXPEDICION", "2019-05-14"); resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want 200", resp.StatusCode)
		}
	})

	t.Run("un error técnico del Registro Civil no consume intento", func(t *testing.T) {
		app, mr, _ := setupBloqueoApp(t)
		for i := 0; i < 3; i++ {
			if resp, _ := postValidarIdentidad(t, app, external.CedulaServicioCaido, "FECHA_EXPEDICION", "2019-05-14"); resp.StatusCode != http.StatusServiceUnavailable {
				t.Fatalf("status = %d; want 503", resp.StatusCode)
			}
		}
		if mr.Exists("srp:dev:auth:reg:attempts:" + external.CedulaServicioCaido) {
			t.Error("los errores técnicos no deben incrementar el contador")
		}
	})

	t.Run("con Redis caído rechaza con 503 en lugar de permitir intentos ilimitados", func(t *testing.T) {
		app, mr, _ := setupBloqueoApp(t)
		mr.Close()
		if resp, _ := postValidarIdentidad(t, app, "0926687856", "FECHA_EXPEDICION", "2019-05-14"); resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("status = %d; want 503", resp.StatusCode)
		}
	})
}

func TestBloqueoLogin(t *testing.T) {
	t.Run("al 3er fallo responde 423 y bloquea incluso la contraseña correcta", func(t *testing.T) {
		app, _, _ := setupBloqueoApp(t)
		for i := 1; i <= 2; i++ {
			if resp, _ := postLogin(t, app, "0926687856", "Incorrecta1*"); resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("intento %d: status = %d; want 401", i, resp.StatusCode)
			}
		}
		resp, jsonRes := postLogin(t, app, "0926687856", "Incorrecta1*")
		esperarBloqueo(t, resp, jsonRes, "AUTH_LOGIN_BLOQUEADO", msgLoginBloqueado)

		if resp, _ := postLogin(t, app, "0926687856", "BiessPostulante2026*"); resp.StatusCode != http.StatusLocked {
			t.Errorf("status con contraseña correcta = %d; want 423", resp.StatusCode)
		}
	})

	t.Run("un login exitoso reinicia el contador", func(t *testing.T) {
		app, _, _ := setupBloqueoApp(t)
		postLogin(t, app, "0926687856", "Incorrecta1*")
		postLogin(t, app, "0926687856", "Incorrecta1*")
		if resp, _ := postLogin(t, app, "0926687856", "BiessPostulante2026*"); resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d; want 200", resp.StatusCode)
		}
		if resp, _ := postLogin(t, app, "0926687856", "Incorrecta1*"); resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("tras un éxito, un fallo no debe bloquear: status = %d", resp.StatusCode)
		}
	})
}
