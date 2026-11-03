package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/external"
	interfacesHTTP "srp-biess/internal/interfaces/http"
	"srp-biess/internal/interfaces/http/handlers"
)

const postulacionDePrueba = "22222222-2222-2222-2222-222222222222"

// fakeCargoRepo simula la persistencia del cargo (SRP-017a) con una sola postulación en memoria.
type fakeCargoRepo struct {
	cedula        string
	estado        *domain.EstadoFormulario
	elegib        string
	escrituras    int
	cargosActivos map[int]domain.SectorCargo
}

func nuevoFakeCargo(cedula string) *fakeCargoRepo {
	return &fakeCargoRepo{
		cedula: cedula,
		elegib: domain.ElegibilidadPendiente,
		estado: &domain.EstadoFormulario{ID: postulacionDePrueba, Estado: "BORRADOR", Version: 3},
		cargosActivos: map[int]domain.SectorCargo{
			1: domain.SectorAfiliados, 2: domain.SectorJubilados, 3: domain.SectorEmpleadores,
		},
	}
}

const postulacionAjena = "33333333-3333-3333-3333-333333333333"

func (f *fakeCargoRepo) ContextoConfirmacion(_ context.Context, id, postulanteID string, cargoID int) (*domain.ContextoCargo, error) {
	if id == postulacionAjena {
		return nil, domain.ErrAccesoDenegado
	}
	if id != postulacionDePrueba || postulanteID != postulanteDePrueba {
		return nil, domain.ErrNotFound
	}
	sector, ok := f.cargosActivos[cargoID]
	if !ok {
		return nil, domain.ErrCargoInvalido
	}
	return &domain.ContextoCargo{
		Estado: f.estado.Estado, Version: f.estado.Version, ElegibilidadEstado: f.elegib,
		Cedula: f.cedula, Sector: sector,
	}, nil
}

func (f *fakeCargoRepo) ConfirmarCargo(_ context.Context, _, _ string, r domain.ResultadoElegibilidad, version int) error {
	if version != f.estado.Version {
		return domain.ErrConflictoVersion
	}
	f.escrituras++
	f.elegib = domain.ElegibilidadNoElegible
	if r.Elegible {
		f.elegib = domain.ElegibilidadElegible
	} else {
		f.estado.Estado = "NO_ELEGIBLE"
	}
	f.estado.Version++
	f.estado.Elegibilidad = &domain.Elegibilidad{Elegible: r.Elegible, CargoID: r.CargoID, Mensaje: r.Motivo, Fecha: time.Now()}
	return nil
}

func (f *fakeCargoRepo) EstadoPostulacion(_ context.Context, _, _ string) (*domain.EstadoFormulario, error) {
	return f.estado, nil
}

// fakeTx ejecuta la función sin transacción real; registra cuántas se abrieron.
type fakeTx struct{ abiertas int }

func (t *fakeTx) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	t.abiertas++
	return fn(ctx)
}

type fakeAuditoria struct{ eventos []domain.EventoAuditoria }

func (a *fakeAuditoria) Registrar(_ context.Context, e domain.EventoAuditoria) error {
	a.eventos = append(a.eventos, e)
	return nil
}

type cargoEnv struct {
	app       *fiber.App
	repo      *fakeCargoRepo
	auditoria *fakeAuditoria
	token     string
}

func setupCargoApp(t *testing.T, cedula string) cargoEnv {
	t.Helper()
	repo := nuevoFakeCargo(cedula)
	auditoria := &fakeAuditoria{}
	authService := application.NewAuthService(nil, nil)
	token, err := authService.GenerarAccessToken(postulanteDePrueba, cedula, "postulante.test@biess.gob.ec")
	if err != nil {
		t.Fatalf("no se pudo generar el token: %v", err)
	}
	elegibilidad := application.NewEligibilityService(external.NewMockEmploymentProvider())
	app := fiber.New()
	interfacesHTTP.SetupRoutes(app, interfacesHTTP.RouterConfig{
		HealthHandler:      handlers.NewHealthHandler(),
		PostulacionHandler: handlers.NewPostulacionHandler(application.NewFormularioService(nuevoFakeRepo())),
		CargoHandler: handlers.NewCargoHandler(
			application.NewCargoService(&fakeTx{}, repo, auditoria, elegibilidad, time.Second)),
		AuthService: authService,
	})
	return cargoEnv{app: app, repo: repo, auditoria: auditoria, token: token}
}

const urlCargo = "/api/v1/postulaciones/" + postulacionDePrueba + "/cargo"

func TestCargoHandler_Elegible(t *testing.T) {
	e := setupCargoApp(t, "0926687856") // afiliado en el mock IESS
	status, out := llamar(t, e.app, http.MethodPut, urlCargo, e.token, map[string]any{"cargoId": 1, "confirmacion": true, "version": 3})
	if status != http.StatusOK {
		t.Fatalf("esperado 200, obtenido %d: %v", status, out)
	}
	data := out["data"].(map[string]any)
	// Campos de SDD 7.5.4.
	if data["elegible"] != true || data["mensaje"] != "Postulante afiliado al IESS" || data["requiereRespaldo"] != true || data["version"] != float64(4) {
		t.Fatalf("contrato SDD incompleto: %v", data)
	}
	// Contrato que consume la SPA.
	elegib, ok := data["elegibilidad"].(map[string]any)
	if !ok || elegib["elegible"] != true || elegib["cargoId"] != float64(1) || elegib["fecha"] == "" {
		t.Fatalf("elegibilidad inesperada: %v", data["elegibilidad"])
	}
	if p := data["postulacion"].(map[string]any); p["estado"] != "BORRADOR" || p["version"] != float64(4) {
		t.Fatalf("postulación inesperada: %v", p)
	}

	if len(e.auditoria.eventos) != 1 {
		t.Fatalf("esperado 1 evento de auditoría, obtenidos %d", len(e.auditoria.eventos))
	}
	ev := e.auditoria.eventos[0]
	if ev.Accion != domain.AccionCargoConfirmado || ev.EntidadID != postulacionDePrueba || ev.ActorID != postulanteDePrueba ||
		ev.Antes["elegibilidadEstado"] != domain.ElegibilidadPendiente || ev.Despues["elegibilidadEstado"] != domain.ElegibilidadElegible {
		t.Fatalf("evento inesperado: %+v", ev)
	}
	if _, conCedula := ev.Despues["cedula"]; conCedula {
		t.Fatalf("la auditoría no debe guardar la cédula")
	}

	// La selección es definitiva: un segundo intento se rechaza.
	if status, out := llamar(t, e.app, http.MethodPut, urlCargo, e.token, map[string]any{"cargoId": 1, "confirmacion": true, "version": 4}); status != http.StatusConflict || out["code"] != "POS_CARGO_CONFIRMADO" {
		t.Fatalf("esperado 409 POS_CARGO_CONFIRMADO, obtenido %d %v", status, out)
	}
}

func TestCargoHandler_NoElegible(t *testing.T) {
	e := setupCargoApp(t, "0926687856") // afiliado, no jubilado
	status, out := llamar(t, e.app, http.MethodPut, urlCargo, e.token, map[string]any{"cargoId": 2, "confirmacion": true, "version": 3})
	mensaje := domain.NewEligibilityError(domain.SectorJubilados).Mensaje
	// SDD 7.5.4: 422 con código, detalle y estado.
	if status != http.StatusUnprocessableEntity || out["code"] != "POS_NO_ELEGIBLE_JUB" || out["detail"] != mensaje || out["estado"] != "NO_ELEGIBLE" {
		t.Fatalf("esperado 422 POS_NO_ELEGIBLE_JUB, obtenido %d %v", status, out)
	}
	data := out["data"].(map[string]any)
	if elegib := data["elegibilidad"].(map[string]any); elegib["elegible"] != false || elegib["mensaje"] != mensaje {
		t.Fatalf("elegibilidad inesperada: %v", elegib)
	}
	if e.repo.estado.Estado != "NO_ELEGIBLE" {
		t.Fatalf("la postulación debía quedar NO_ELEGIBLE")
	}
	if len(e.auditoria.eventos) != 1 || e.auditoria.eventos[0].Accion != domain.AccionPostulacionNoElegible ||
		e.auditoria.eventos[0].Despues["estado"] != "NO_ELEGIBLE" {
		t.Fatalf("el cierre debía quedar auditado: %+v", e.auditoria.eventos)
	}
}

func TestCargoHandler_ProveedorCaido(t *testing.T) {
	e := setupCargoApp(t, external.CedulaServicioCaido)
	body := map[string]any{"cargoId": 1, "confirmacion": true, "version": 3}

	status, out := llamar(t, e.app, http.MethodPut, urlCargo, e.token, body)
	if status != http.StatusServiceUnavailable || out["code"] != "INTEGRACION_NO_DISPONIBLE" {
		t.Fatalf("esperado 503 INTEGRACION_NO_DISPONIBLE, obtenido %d %v", status, out)
	}
	if e.repo.escrituras != 0 || e.repo.elegib != domain.ElegibilidadPendiente || e.repo.estado.Version != 3 || len(e.auditoria.eventos) != 0 {
		t.Fatalf("no debía escribirse nada con el proveedor caído: %+v", e.repo)
	}

	// El reintento queda habilitado: con el proveedor recuperado se confirma.
	e.repo.cedula = "0926687856"
	if status, out := llamar(t, e.app, http.MethodPut, urlCargo, e.token, body); status != http.StatusOK {
		t.Fatalf("el reintento debía responder 200, obtenido %d %v", status, out)
	}
}

func TestCargoHandler_Rechazos(t *testing.T) {
	cuerpo := map[string]any{"cargoId": 1, "confirmacion": true, "version": 3}
	casos := []struct {
		nombre string
		url    string
		body   map[string]any
		status int
		code   string
	}{
		{"cargo inexistente", urlCargo, map[string]any{"cargoId": 9, "confirmacion": true, "version": 3}, http.StatusBadRequest, "POS_CARGO_INVALIDO"},
		{"sin confirmación", urlCargo, map[string]any{"cargoId": 1, "confirmacion": false, "version": 3}, http.StatusBadRequest, "VALIDATION_ERROR"},
		{"sin versión", urlCargo, map[string]any{"cargoId": 1, "confirmacion": true}, http.StatusBadRequest, "VALIDATION_ERROR"},
		{"versión desactualizada", urlCargo, map[string]any{"cargoId": 1, "confirmacion": true, "version": 2}, http.StatusConflict, "POS_CONFLICTO_VERSION"},
		{"postulación inexistente", "/api/v1/postulaciones/44444444-4444-4444-4444-444444444444/cargo", cuerpo, http.StatusNotFound, "POS_NO_ENCONTRADA"},
		{"postulación ajena", "/api/v1/postulaciones/" + postulacionAjena + "/cargo", cuerpo, http.StatusForbidden, "POS_ACCESO_DENEGADO"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := setupCargoApp(t, "0926687856")
			status, out := llamar(t, e.app, http.MethodPut, c.url, e.token, c.body)
			if status != c.status || out["code"] != c.code {
				t.Fatalf("esperado %d %s, obtenido %d %v", c.status, c.code, status, out)
			}
			if e.repo.escrituras != 0 {
				t.Fatalf("un rechazo no debía escribir")
			}
			// SDD 9.9: solo el acceso denegado se audita.
			auditados := len(e.auditoria.eventos)
			if c.status == http.StatusForbidden {
				if auditados != 1 || e.auditoria.eventos[0].Accion != domain.AccionAccesoDenegado || e.auditoria.eventos[0].EntidadID != postulacionAjena {
					t.Fatalf("el acceso denegado debía auditarse: %+v", e.auditoria.eventos)
				}
			} else if auditados != 0 {
				t.Fatalf("un rechazo de validación no debía auditarse: %+v", e.auditoria.eventos)
			}
		})
	}

	t.Run("sin token responde 401", func(t *testing.T) {
		e := setupCargoApp(t, "0926687856")
		if status, _ := llamar(t, e.app, http.MethodPut, urlCargo, "", map[string]any{}); status != http.StatusUnauthorized {
			t.Fatalf("esperado 401, obtenido %d", status)
		}
	})

	t.Run("postulación cerrada responde 409", func(t *testing.T) {
		e := setupCargoApp(t, "0926687856")
		e.repo.estado.Estado = "ENVIADA"
		status, out := llamar(t, e.app, http.MethodPut, urlCargo, e.token, cuerpo)
		if status != http.StatusConflict || out["code"] != "POS_CERRADA" {
			t.Fatalf("esperado 409 POS_CERRADA, obtenido %d %v", status, out)
		}
	})
}
