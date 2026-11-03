package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/infrastructure/database"
	interfacesHTTP "srp-biess/internal/interfaces/http"
	"srp-biess/internal/interfaces/http/handlers"
)

func nombres(body map[string]any) []string {
	items, _ := body["data"].([]any)
	out := make([]string, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			nombre, _ := m["nombre"].(string)
			out = append(out, nombre)
		}
	}
	return out
}

func contiene(lista []string, valor string) bool {
	for _, v := range lista {
		if v == valor {
			return true
		}
	}
	return false
}

// Integración con el seed de SDD 16.3: se omite si PostgreSQL no está disponible.
func TestCatalogoHandler_Cascada(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	defer pool.Close()

	app := fiber.New()
	interfacesHTTP.SetupRoutes(app, interfacesHTTP.RouterConfig{
		HealthHandler:   handlers.NewHealthHandler(),
		CatalogoHandler: handlers.NewCatalogoHandler(application.NewCatalogoService(database.NewPostgresCatalogoRepository(pool))),
	})

	status, body := llamar(t, app, http.MethodGet, "/api/v1/catalogos/paises", "", nil)
	if status != http.StatusOK || !contiene(nombres(body), "Ecuador") {
		t.Fatalf("países inesperados: %d %v", status, body)
	}

	_, body = llamar(t, app, http.MethodGet, "/api/v1/catalogos/paises/1/provincias", "", nil)
	if provincias := nombres(body); len(provincias) != 24 || !contiene(provincias, "Pichincha") {
		t.Fatalf("se esperaban las 24 provincias del Ecuador, obtenido %v", provincias)
	}

	_, body = llamar(t, app, http.MethodGet, "/api/v1/catalogos/provincias/17/cantones", "", nil)
	cantones := nombres(body)
	if !contiene(cantones, "Quito") || contiene(cantones, "Guayaquil") {
		t.Fatalf("solo debían llegar cantones de Pichincha, obtenido %v", cantones)
	}

	_, body = llamar(t, app, http.MethodGet, "/api/v1/catalogos/provincias/2/cantones", "", nil)
	if data, ok := body["data"].([]any); !ok || len(data) != 0 {
		t.Fatalf("una provincia sin cantones debe devolver [], obtenido %v", body["data"])
	}

	if status, _ := llamar(t, app, http.MethodGet, "/api/v1/catalogos/paises/abc/provincias", "", nil); status != http.StatusBadRequest {
		t.Fatalf("un id no numérico debía responder 400, obtenido %d", status)
	}
}

func TestCatalogoHandler_SinPersistencia(t *testing.T) {
	app := fiber.New()
	interfacesHTTP.SetupRoutes(app, interfacesHTTP.RouterConfig{
		HealthHandler:   handlers.NewHealthHandler(),
		CatalogoHandler: handlers.NewCatalogoHandler(nil),
	})
	if status, _ := llamar(t, app, http.MethodGet, "/api/v1/catalogos/paises", "", nil); status != http.StatusServiceUnavailable {
		t.Fatalf("esperado 503 sin PostgreSQL, obtenido %d", status)
	}
}
