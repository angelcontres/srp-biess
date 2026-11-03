package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
	"srp-biess/internal/infrastructure/security"
	"srp-biess/internal/infrastructure/storage"
	interfacesHTTP "srp-biess/internal/interfaces/http"
	"srp-biess/internal/interfaces/http/handlers"
)

func multipartPDF(t *testing.T, tipo, nombre string, contenido []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if tipo != "" {
		_ = w.WriteField("tipo", tipo)
	}
	if nombre != "" {
		part, _ := w.CreateFormFile("archivo", nombre)
		_, _ = part.Write(contenido)
	}
	_ = w.Close()
	return &buf, w.FormDataContentType()
}

func peticion(t *testing.T, app *fiber.App, req *http.Request, token string) (*http.Response, map[string]any, []byte) {
	t.Helper()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, 10000)
	if err != nil {
		t.Fatalf("error en la petición: %v", err)
	}
	cuerpo, _ := io.ReadAll(resp.Body)
	var out map[string]any
	_ = json.Unmarshal(cuerpo, &out)
	return resp, out, cuerpo
}

// Contrato HTTP de SRP-017b contra PostgreSQL y el sistema de archivos.
func TestDocumentoHandler_Integracion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}

	sufijo := time.Now().UnixNano()
	p := &domain.Postulante{
		Cedula: fmt.Sprintf("06%08d", sufijo%100000000), Nombres: "PRUEBA", Apellidos: "HTTP DOCUMENTOS",
		Email: fmt.Sprintf("http_docs_%d@biess.gob.ec", sufijo), PasswordHash: "hash_seguro",
	}
	if err := database.NewPostgresPostulanteRepository(pool).Save(ctx, p); err != nil {
		t.Fatalf("no se pudo crear el postulante: %v", err)
	}
	t.Cleanup(func() {
		bg := context.Background()
		_, _ = pool.Exec(bg, `DELETE FROM auditoria_evento WHERE actor_id = $1::uuid`, p.ID)
		_, _ = pool.Exec(bg, `DELETE FROM documento WHERE postulante_id = $1::uuid`, p.ID)
		_, _ = pool.Exec(bg, `DELETE FROM postulacion WHERE postulante_id = $1::uuid`, p.ID)
		_, _ = pool.Exec(bg, `DELETE FROM postulante WHERE id = $1::uuid`, p.ID)
	})
	borrador, err := database.NewPostgresFormularioRepository(pool).CrearBorrador(ctx, p.ID)
	if errors.Is(err, domain.ErrFaseCerrada) {
		t.Skip("No hay convocatoria con la fase habilitada en la fecha actual")
	}
	if err != nil {
		t.Fatalf("CrearBorrador: %v", err)
	}

	almacen, err := storage.NewAlmacenLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	authService := application.NewAuthService(nil, nil)
	token, _ := authService.GenerarAccessToken(p.ID, p.Cedula, p.Email)
	svc := application.NewDocumentoService(database.NewPgxTxManager(pool), database.NewPostgresDocumentoRepository(pool), almacen,
		security.NewFirmadorHMAC([]byte("secreto-http")), database.NewPostgresAuditoriaRepository(pool), domain.LimitesPorDefecto, time.Minute)
	app := fiber.New(fiber.Config{BodyLimit: 5 * 1024 * 1024})
	interfacesHTTP.SetupRoutes(app, interfacesHTTP.RouterConfig{
		HealthHandler:      handlers.NewHealthHandler(),
		PostulacionHandler: handlers.NewPostulacionHandler(application.NewFormularioService(nuevoFakeRepo())),
		DocumentoHandler:   handlers.NewDocumentoHandler(svc),
		AuthService:        authService,
	})
	base := "/api/v1/postulaciones/" + borrador.ID + "/documentos"
	subir := func(tipo, nombre string, contenido []byte) (*http.Response, map[string]any) {
		body, ct := multipartPDF(t, tipo, nombre, contenido)
		req := httptest.NewRequest(http.MethodPost, base, body)
		req.Header.Set("Content-Type", ct)
		resp, out, _ := peticion(t, app, req, token)
		return resp, out
	}
	contenido := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 2048)...)

	resp, out := subir("TITULO", "Título Ingeniería.pdf", contenido)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperado 201, obtenido %d %v", resp.StatusCode, out)
	}
	doc := out["data"].(map[string]any)
	for _, campo := range []string{"id", "nombreOriginal", "tamanoBytes", "hashSha256", "fechaCarga", "urlDescarga"} {
		if doc[campo] == nil || doc[campo] == "" {
			t.Fatalf("falta %s en la respuesta: %v", campo, doc)
		}
	}
	if doc["nombreOriginal"] != "Título Ingeniería.pdf" || doc["tamanoBytes"] != float64(len(contenido)) {
		t.Fatalf("respuesta inesperada: %v", doc)
	}
	urlDescarga := doc["urlDescarga"].(string)
	if !strings.HasPrefix(urlDescarga, base+"/"+doc["id"].(string)+"/download?exp=") {
		t.Fatalf("URL de descarga inesperada: %s", urlDescarga)
	}

	t.Run("rechazos con el código del catálogo", func(t *testing.T) {
		casos := []struct {
			nombre, tipo, archivo string
			contenido             []byte
			status                int
			code                  string
		}{
			{"no es PDF", "TITULO", "foto.pdf", []byte("\xFF\xD8\xFF\xE0"), http.StatusUnsupportedMediaType, "DOC_TIPO_INVALIDO"},
			{"excede 2 MB", "TITULO", "t.pdf", append([]byte("%PDF-"), bytes.Repeat([]byte("x"), 2*1024*1024)...), http.StatusRequestEntityTooLarge, "DOC_TAMANO_EXCEDIDO"},
			{"FORMULARIO_POSTULACION", "FORMULARIO_POSTULACION", "f.pdf", contenido, http.StatusBadRequest, "VALIDATION_ERROR"},
			{"sin archivo", "TITULO", "", nil, http.StatusBadRequest, "VALIDATION_ERROR"},
		}
		for _, c := range casos {
			resp, out := subir(c.tipo, c.archivo, c.contenido)
			if resp.StatusCode != c.status || out["code"] != c.code {
				t.Errorf("%s: esperado %d %s, obtenido %d %v", c.nombre, c.status, c.code, resp.StatusCode, out)
			}
		}
	})

	t.Run("descarga con JWT y firma (doble validación)", func(t *testing.T) {
		resp, _, cuerpo := peticion(t, app, httptest.NewRequest(http.MethodGet, urlDescarga, nil), token)
		if resp.StatusCode != http.StatusOK || !bytes.Equal(cuerpo, contenido) {
			t.Fatalf("esperado 200 con el PDF, obtenido %d (%d bytes)", resp.StatusCode, len(cuerpo))
		}
		if resp.Header.Get("Content-Type") != "application/pdf" || resp.Header.Get("X-Content-Type-Options") != "nosniff" ||
			!strings.Contains(resp.Header.Get("Content-Disposition"), "attachment") ||
			!strings.Contains(resp.Header.Get("Content-Disposition"), "T%C3%ADtulo%20Ingenier%C3%ADa.pdf") {
			t.Fatalf("cabeceras inesperadas: %v", resp.Header)
		}
		if resp, _, _ := peticion(t, app, httptest.NewRequest(http.MethodGet, urlDescarga, nil), ""); resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("sin JWT: esperado 401, obtenido %d", resp.StatusCode)
		}
		alterada := urlDescarga[:len(urlDescarga)-2] + "00"
		if resp, out, _ := peticion(t, app, httptest.NewRequest(http.MethodGet, alterada, nil), token); resp.StatusCode != http.StatusForbidden || out["code"] != "DOC_ENLACE_INVALIDO" {
			t.Fatalf("firma alterada: esperado 403 DOC_ENLACE_INVALIDO, obtenido %d %v", resp.StatusCode, out)
		}
		resp, out, _ = peticion(t, app, httptest.NewRequest(http.MethodGet, base+"/"+doc["id"].(string), nil), token)
		if resp.StatusCode != http.StatusOK || out["data"].(map[string]any)["urlDescarga"] == "" {
			t.Fatalf("metadatos: esperado 200 con una URL nueva, obtenido %d %v", resp.StatusCode, out)
		}
	})

	t.Run("borrado", func(t *testing.T) {
		url := base + "/" + doc["id"].(string)
		if resp, _, _ := peticion(t, app, httptest.NewRequest(http.MethodDelete, url, nil), token); resp.StatusCode != http.StatusNoContent {
			t.Fatalf("esperado 204, obtenido %d", resp.StatusCode)
		}
		if resp, out, _ := peticion(t, app, httptest.NewRequest(http.MethodDelete, url, nil), token); resp.StatusCode != http.StatusNotFound || out["code"] != "DOC_NO_ENCONTRADO" {
			t.Fatalf("segundo borrado: esperado 404 DOC_NO_ENCONTRADO, obtenido %d %v", resp.StatusCode, out)
		}
	})
}
