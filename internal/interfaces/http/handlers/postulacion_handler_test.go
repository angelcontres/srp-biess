package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	interfacesHTTP "srp-biess/internal/interfaces/http"
	"srp-biess/internal/interfaces/http/handlers"
)

const postulanteDePrueba = "11111111-1111-1111-1111-111111111111"

// fakeFormularioRepo simula PostgreSQL en memoria para probar el contrato HTTP (SRP-013).
type fakeFormularioRepo struct {
	estado    *domain.EstadoFormulario
	secciones map[string][]byte
	completas map[string]bool
}

func nuevoFakeRepo() *fakeFormularioRepo {
	return &fakeFormularioRepo{secciones: map[string][]byte{}, completas: map[string]bool{}}
}

func (f *fakeFormularioRepo) ObtenerActual(_ context.Context, postulanteID string) (*domain.EstadoFormulario, error) {
	if f.estado == nil || postulanteID != postulanteDePrueba {
		return nil, domain.ErrNotFound
	}
	return f.estado, nil
}

func (f *fakeFormularioRepo) CrearBorrador(_ context.Context, _ string) (*domain.EstadoFormulario, error) {
	f.estado = &domain.EstadoFormulario{
		ID: "22222222-2222-2222-2222-222222222222", Estado: "BORRADOR", Version: 1,
		Secciones: map[string]domain.EstadoSeccion{}, SeccionReanudacion: "personal",
	}
	return f.estado, nil
}

func (f *fakeFormularioRepo) ObtenerDetalle(_ context.Context, id, postulanteID string) (*domain.DetalleFormulario, error) {
	if f.estado == nil || id != f.estado.ID || postulanteID != postulanteDePrueba {
		return nil, domain.ErrNotFound
	}
	return &domain.DetalleFormulario{
		Postulacion: f.estado,
		Verificados: &domain.DatosVerificados{
			Cedula:       "0926687856",
			Nacionalidad: "ECUATORIANA",
			Correo:       "postulante.test@biess.gob.ec",
		},
		Secciones: f.secciones,
	}, nil
}

func (f *fakeFormularioRepo) GuardarSeccion(_ context.Context, id, _ string, seccion string, datos []byte, completa bool, version int) (*domain.EstadoFormulario, error) {
	if f.estado == nil || id != f.estado.ID {
		return nil, domain.ErrNotFound
	}
	if f.estado.Estado != "BORRADOR" {
		return nil, domain.ErrPostulacionCerrada
	}
	if version != f.estado.Version {
		return nil, domain.ErrConflictoVersion
	}
	f.secciones[seccion] = datos
	f.completas[seccion] = completa
	f.estado.Version++
	return f.estado, nil
}

func setupPostulacionApp(t *testing.T) (*fiber.App, *fakeFormularioRepo, string) {
	t.Helper()
	repo := nuevoFakeRepo()
	authService := application.NewAuthService(nil, nil)
	token, err := authService.GenerarAccessToken(postulanteDePrueba, "0926687856", "postulante.test@biess.gob.ec")
	if err != nil {
		t.Fatalf("no se pudo generar el token: %v", err)
	}
	app := fiber.New()
	interfacesHTTP.SetupRoutes(app, interfacesHTTP.RouterConfig{
		HealthHandler:      handlers.NewHealthHandler(),
		PostulacionHandler: handlers.NewPostulacionHandler(application.NewFormularioService(repo)),
		AuthService:        authService,
	})
	return app, repo, token
}

func llamar(t *testing.T, app *fiber.App, method, url, token string, body any) (int, map[string]any) {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, url, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en la petición: %v", err)
	}
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func TestPostulacionHandler_Flujo(t *testing.T) {
	app, repo, token := setupPostulacionApp(t)
	base := "/api/v1/postulaciones"

	t.Run("sin token responde 401", func(t *testing.T) {
		if status, _ := llamar(t, app, http.MethodGet, base+"/actual", "", nil); status != http.StatusUnauthorized {
			t.Fatalf("esperado 401, obtenido %d", status)
		}
	})

	t.Run("sin borrador responde 404 POS_NO_ENCONTRADA", func(t *testing.T) {
		status, body := llamar(t, app, http.MethodGet, base+"/actual", token, nil)
		if status != http.StatusNotFound || body["code"] != "POS_NO_ENCONTRADA" {
			t.Fatalf("esperado 404 POS_NO_ENCONTRADA, obtenido %d %v", status, body)
		}
	})

	t.Run("crea el borrador una sola vez", func(t *testing.T) {
		if status, _ := llamar(t, app, http.MethodPost, base, token, nil); status != http.StatusCreated {
			t.Fatalf("esperado 201, obtenido %d", status)
		}
		status, body := llamar(t, app, http.MethodPost, base, token, nil)
		if status != http.StatusConflict || body["code"] != "POS_YA_EXISTE" {
			t.Fatalf("esperado 409 POS_YA_EXISTE, obtenido %d %v", status, body)
		}
	})

	id := repo.estado.ID

	t.Run("RF-EF12-02: permite guardar borrador con datos parciales (completa=false)", func(t *testing.T) {
		status, body := llamar(t, app, http.MethodPut, base+"/"+id+"/domicilio", token,
			map[string]any{"callePrincipal": "Av. Amazonas", "version": 1, "completa": false})
		if status != http.StatusOK {
			t.Fatalf("esperado 200 al guardar borrador, obtenido %d %v", status, body)
		}
		var guardado map[string]any
		_ = json.Unmarshal(repo.secciones["domicilio"], &guardado)
		if guardado["callePrincipal"] != "Av. Amazonas" || guardado["version"] != nil || guardado["completa"] != nil {
			t.Fatalf("datos guardados inesperados: %v", guardado)
		}
		if repo.completas["domicilio"] {
			t.Fatal("el borrador no debía marcarse como completa")
		}
		if repo.estado.Version != 2 {
			t.Fatalf("esperada versión 2, obtenida %d", repo.estado.Version)
		}
	})

	t.Run("RF-EF12-24: rechaza marcar COMPLETA cuando faltan campos requeridos (HTTP 422 VALIDATION_ERROR)", func(t *testing.T) {
		status, body := llamar(t, app, http.MethodPut, base+"/"+id+"/domicilio", token,
			map[string]any{"callePrincipal": "Av. Amazonas", "version": 2, "completa": true})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("esperado 422, obtenido %d %v", status, body)
		}
		if body["code"] != "VALIDATION_ERROR" {
			t.Fatalf("esperado code VALIDATION_ERROR, obtenido %v", body["code"])
		}
		detalles, ok := body["detalles"].([]any)
		if !ok || len(detalles) == 0 {
			t.Fatalf("esperados detalles de error, obtenido %v", body)
		}
		if repo.completas["domicilio"] {
			t.Fatal("no debía marcarse como completa en fallo de validación")
		}
		if repo.estado.Version != 2 {
			t.Fatalf("la versión no debió incrementar tras error 422, actual: %d", repo.estado.Version)
		}
	})

	t.Run("RF-EF12-24: marca COMPLETA cuando los campos están completos y válidos (HTTP 200)", func(t *testing.T) {
		datosValidos := map[string]any{
			"paisId":          "EC",
			"provinciaId":     "17",
			"cantonId":        "1701",
			"telefonoCelular": "0991234567",
			"callePrincipal":  "Av. Amazonas",
			"numero":          "N24-50",
			"version":         2,
			"completa":        true,
		}
		status, body := llamar(t, app, http.MethodPut, base+"/"+id+"/domicilio", token, datosValidos)
		if status != http.StatusOK {
			t.Fatalf("esperado 200, obtenido %d %v", status, body)
		}
		if !repo.completas["domicilio"] {
			t.Fatal("la sección debió marcarse como completa")
		}
		if repo.estado.Version != 3 {
			t.Fatalf("esperada versión 3, obtenida %d", repo.estado.Version)
		}
	})

	t.Run("versión desactualizada responde 409 POS_CONFLICTO_VERSION", func(t *testing.T) {
		status, body := llamar(t, app, http.MethodPatch, base+"/"+id+"/informacion-personal", token,
			map[string]any{"estadoCivil": "SOLTERO", "version": 2})
		if status != http.StatusConflict || body["code"] != "POS_CONFLICTO_VERSION" {
			t.Fatalf("esperado 409 POS_CONFLICTO_VERSION, obtenido %d %v", status, body)
		}
	})

	t.Run("sin versión responde 400", func(t *testing.T) {
		if status, _ := llamar(t, app, http.MethodPut, base+"/"+id+"/familiares", token, map[string]any{}); status != http.StatusBadRequest {
			t.Fatalf("esperado 400, obtenido %d", status)
		}
	})

	t.Run("postulación en estado cerrado (no BORRADOR) responde 409 POS_CERRADA", func(t *testing.T) {
		repo.estado.Estado = "ENVIADA"
		defer func() { repo.estado.Estado = "BORRADOR" }()

		status, body := llamar(t, app, http.MethodPatch, base+"/"+id+"/informacion-personal", token,
			map[string]any{"estadoCivil": "SOLTERO", "version": 3, "completa": true})
		if status != http.StatusConflict || body["code"] != "POS_CERRADA" {
			t.Fatalf("esperado 409 POS_CERRADA, obtenido %d %v", status, body)
		}
	})

	t.Run("el detalle devuelve verificados y valores guardados", func(t *testing.T) {
		status, body := llamar(t, app, http.MethodGet, base+"/"+id, token, nil)
		if status != http.StatusOK {
			t.Fatalf("esperado 200, obtenido %d", status)
		}
		data, _ := body["data"].(map[string]any)
		secciones, _ := data["secciones"].(map[string]any)
		domicilio, _ := secciones["domicilio"].(map[string]any)
		if domicilio["callePrincipal"] != "Av. Amazonas" || data["verificados"] == nil {
			t.Fatalf("detalle inesperado: %v", data)
		}
	})

	t.Run("ruta inexistente bajo /postulaciones responde 404, no 401", func(t *testing.T) {
		if status, _ := llamar(t, app, http.MethodGet, base+"/"+id+"/no-existe", token, nil); status != http.StatusNotFound {
			t.Fatalf("esperado 404, obtenido %d", status)
		}
	})
}

func TestPostulacionHandler_ValidacionSecciones(t *testing.T) {
	app, repo, token := setupPostulacionApp(t)
	base := "/api/v1/postulaciones"

	// Crear borrador inicial
	status, _ := llamar(t, app, http.MethodPost, base, token, nil)
	if status != http.StatusCreated {
		t.Fatalf("no se pudo crear borrador: %d", status)
	}
	id := repo.estado.ID

	// 1. Información Personal
	t.Run("sección personal incompleta responde 422", func(t *testing.T) {
		status, body := llamar(t, app, http.MethodPatch, base+"/"+id+"/informacion-personal", token,
			map[string]any{"version": repo.estado.Version, "completa": true})
		if status != http.StatusUnprocessableEntity || body["code"] != "VALIDATION_ERROR" {
			t.Fatalf("esperado 422 VALIDATION_ERROR, obtenido %d %v", status, body)
		}
	})

	t.Run("sección personal completa responde 200 y marca completa", func(t *testing.T) {
		status, _ := llamar(t, app, http.MethodPatch, base+"/"+id+"/informacion-personal", token,
			map[string]any{"estadoCivil": "SOLTERO", "version": repo.estado.Version, "completa": true})
		if status != http.StatusOK || !repo.completas["personal"] {
			t.Fatalf("esperado 200 OK y completa true, obtenido %d", status)
		}
	})

	// 2. Familiares
	t.Run("sección familiares responde 422 si falta pregunta", func(t *testing.T) {
		status, body := llamar(t, app, http.MethodPut, base+"/"+id+"/familiares", token,
			map[string]any{"version": repo.estado.Version, "completa": true})
		if status != http.StatusUnprocessableEntity || body["code"] != "VALIDATION_ERROR" {
			t.Fatalf("esperado 422, obtenido %d %v", status, body)
		}
	})

	t.Run("sección familiares completa responde 200", func(t *testing.T) {
		status, _ := llamar(t, app, http.MethodPut, base+"/"+id+"/familiares", token,
			map[string]any{"tieneFamiliares": "NO", "familiares": []any{}, "version": repo.estado.Version, "completa": true})
		if status != http.StatusOK || !repo.completas["familiares"] {
			t.Fatalf("esperado 200 OK, obtenido %d", status)
		}
	})

	// 3. Respaldo organizacional (cargo)
	t.Run("sección cargo incompleta responde 422", func(t *testing.T) {
		status, _ := llamar(t, app, http.MethodPut, base+"/"+id+"/respaldo-organizacional", token,
			map[string]any{"tieneRespaldo": "SI", "version": repo.estado.Version, "completa": true})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("esperado 422, obtenido %d", status)
		}
	})

	t.Run("sección cargo completa responde 200", func(t *testing.T) {
		status, _ := llamar(t, app, http.MethodPut, base+"/"+id+"/respaldo-organizacional", token,
			map[string]any{"tieneRespaldo": "NO", "version": repo.estado.Version, "completa": true})
		if status != http.StatusOK || !repo.completas["cargo"] {
			t.Fatalf("esperado 200 OK, obtenido %d", status)
		}
	})

	// 4. Formación académica
	t.Run("sección formacion vacía responde 422", func(t *testing.T) {
		status, _ := llamar(t, app, http.MethodPut, base+"/"+id+"/formacion-academica", token,
			map[string]any{"formaciones": []any{}, "version": repo.estado.Version, "completa": true})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("esperado 422, obtenido %d", status)
		}
	})

	// 5. Inhabilidades
	t.Run("sección inhabilidades incompleta responde 422", func(t *testing.T) {
		status, _ := llamar(t, app, http.MethodPut, base+"/"+id+"/inhabilidades", token,
			map[string]any{"respuestas": []string{"NO"}, "version": repo.estado.Version, "completa": true})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("esperado 422, obtenido %d", status)
		}
	})
}

func TestFormularioService_SinPersistencia(t *testing.T) {
	_, err := application.NewFormularioService(nil).ObtenerActual(context.Background(), postulanteDePrueba)
	if err != application.ErrPersistenciaNoDisponible {
		t.Fatalf("esperado ErrPersistenciaNoDisponible, obtenido %v", err)
	}
}
