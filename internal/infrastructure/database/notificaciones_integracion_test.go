package database_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
	"srp-biess/internal/infrastructure/email"
	pdfgen "srp-biess/internal/infrastructure/pdf"
	"srp-biess/internal/infrastructure/storage"
)

const mailpitAPI = "http://localhost:8025/api/v1"

// Flujo completo (ADR-011, RF-EF12-23): enviar la postulación encola el correo; el worker lo
// envía por SMTP a Mailpit con el formulario PDF adjunto y lo marca ENVIADA. Requiere PostgreSQL y
// Mailpit (docker compose up -d); se omite si alguno no está disponible.
func TestNotificaciones_EnvioRealAMailpit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if resp, err := http.Get(mailpitAPI + "/info"); err != nil {
		t.Skipf("Mailpit no disponible: %v", err)
	} else {
		resp.Body.Close()
	}
	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	aplicarMigraciones(ctx, t, func(ctx context.Context, sql string) error {
		_, err := pool.Exec(ctx, sql)
		return err
	})

	sufijo := time.Now().UnixNano()
	correo := fmt.Sprintf("mailpit_%d@correo.ec", sufijo)
	p := &domain.Postulante{
		Cedula: fmt.Sprintf("02%08d", sufijo%100000000), Nombres: "MARÍA JOSÉ", Apellidos: "PEÑAFIEL",
		Email: correo, PasswordHash: "hash_seguro",
	}
	postulanteRepo := database.NewPostgresPostulanteRepository(pool)
	if err := postulanteRepo.Save(ctx, p); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		bg := context.Background()
		_, _ = pool.Exec(bg, `DELETE FROM notificacion WHERE postulante_id = $1::uuid`, p.ID)
		_, _ = pool.Exec(bg, `DELETE FROM documento WHERE postulante_id = $1::uuid`, p.ID)
		_, _ = pool.Exec(bg, `DELETE FROM postulacion WHERE postulante_id = $1::uuid`, p.ID)
		_, _ = pool.Exec(bg, `DELETE FROM postulante WHERE id = $1::uuid`, p.ID)
	})
	formularioRepo := database.NewPostgresFormularioRepository(pool)
	borrador, err := formularioRepo.CrearBorrador(ctx, p.ID)
	if errors.Is(err, domain.ErrFaseCerrada) {
		t.Skip("No hay convocatoria con la fase habilitada en la fecha actual")
	}
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE postulacion SET cargo_id = 1, elegibilidad_estado = 'ELEGIBLE', estado = 'LISTA_PARA_ENVIO' WHERE id = $1::uuid`, borrador.ID); err != nil {
		t.Fatal(err)
	}

	almacen, err := storage.NewAlmacenLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx := database.NewPgxTxManager(pool)
	documentoRepo := database.NewPostgresDocumentoRepository(pool)
	notificacionRepo := database.NewPostgresNotificacionRepository(pool)
	envio := application.NewPostulacionService(tx, database.NewPostgresPostulacionRepository(pool), documentoRepo, notificacionRepo).
		WithFormularioPDF(postulanteRepo, formularioRepo, pdfgen.NewGeneradorGopdf(), almacen)
	enviada, err := envio.EnviarPostulacion(ctx, application.EnviarPostulacionRequest{
		PostulacionID: borrador.ID, PostulanteID: p.ID, AceptaConsentimiento: true,
		VersionTerminos: "2026.1", IdempotencyKey: fmt.Sprintf("mailpit-%d", sufijo),
	})
	if err != nil {
		t.Fatalf("EnviarPostulacion: %v", err)
	}

	plantillas, err := email.NewPlantillas()
	if err != nil {
		t.Fatal(err)
	}
	worker := application.NewNotificacionService(tx, notificacionRepo,
		email.NewSMTPSender(email.ConfigSMTP{Host: "localhost", Puerto: 1025, De: "notificaciones-concurso@biess.gob.ec", NombreDe: "Concurso Directorio BIESS"}),
		plantillas, documentoRepo, almacen, database.NewPostgresCatalogoRepository(pool))
	if _, err := worker.ProcesarPendientes(ctx, 50); err != nil {
		t.Fatalf("ProcesarPendientes: %v", err)
	}

	var estado string
	var intentos int
	if err := pool.QueryRow(ctx, `SELECT estado::text, intentos FROM notificacion WHERE postulacion_id = $1::uuid`, borrador.ID).Scan(&estado, &intentos); err != nil {
		t.Fatal(err)
	}
	if estado != "ENVIADA" || intentos != 1 {
		t.Fatalf("notificación: estado=%s intentos=%d", estado, intentos)
	}
	// Procesar otra vez no la reenvía.
	if n, _ := worker.ProcesarPendientes(ctx, 50); n != 0 {
		var pendientes int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM notificacion WHERE postulacion_id = $1::uuid AND estado <> 'ENVIADA'`, borrador.ID).Scan(&pendientes)
		if pendientes != 0 {
			t.Fatalf("la notificación no debía volver a procesarse")
		}
	}

	// El correo llegó a Mailpit con el código y el PDF adjunto.
	var resultado struct {
		Messages []struct {
			ID          string
			Subject     string
			Attachments int
		} `json:"messages"`
	}
	resp, err := http.Get(mailpitAPI + "/search?query=" + url.QueryEscape("to:"+correo))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&resultado); err != nil {
		t.Fatal(err)
	}
	if len(resultado.Messages) != 1 {
		t.Fatalf("esperado 1 correo en Mailpit para %s, hay %d", correo, len(resultado.Messages))
	}
	m := resultado.Messages[0]
	if !strings.Contains(m.Subject, enviada.CodigoUnico) || m.Attachments != 1 {
		t.Fatalf("correo inesperado: %+v", m)
	}
	t.Cleanup(func() {
		req, _ := http.NewRequest(http.MethodDelete, mailpitAPI+"/messages", strings.NewReader(`{"IDs":["`+m.ID+`"]}`))
		req.Header.Set("Content-Type", "application/json")
		if r, err := http.DefaultClient.Do(req); err == nil {
			r.Body.Close()
		}
	})
}
