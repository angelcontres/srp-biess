package database_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
	pdfgen "srp-biess/internal/infrastructure/pdf"
	"srp-biess/internal/infrastructure/storage"
)

// Flujo SDD 7.5.6 + 7.5.7: BORRADOR → (guardar) LISTA_PARA_ENVIO → (enviar) ENVIADA.
func TestGuardarParaEnvio_Integracion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
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
	p := &domain.Postulante{
		Cedula: fmt.Sprintf("04%08d", sufijo%100000000), Nombres: "LUCÍA", Apellidos: "GUARDAR ENVÍO",
		Email: fmt.Sprintf("guardar_%d@biess.gob.ec", sufijo), PasswordHash: "hash_seguro",
	}
	postulanteRepo := database.NewPostgresPostulanteRepository(pool)
	if err := postulanteRepo.Save(ctx, p); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		bg := context.Background()
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

	almacen, err := storage.NewAlmacenLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	postulacionRepo := database.NewPostgresPostulacionRepository(pool)
	svc := application.NewPostulacionService(database.NewPgxTxManager(pool), postulacionRepo,
		database.NewPostgresDocumentoRepository(pool), database.NewPostgresNotificacionRepository(pool)).
		WithFormularioPDF(postulanteRepo, formularioRepo, pdfgen.NewGeneradorGopdf(), almacen).
		WithAuditoria(database.NewPostgresAuditoriaRepository(pool))
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM auditoria_evento WHERE entidad_id = $1`, borrador.ID)
	})
	guardar := func(postulanteID string) (*application.GuardarParaEnvioResult, error) {
		return svc.GuardarParaEnvio(ctx, application.GuardarParaEnvioRequest{
			PostulacionID: borrador.ID, PostulanteID: postulanteID, AceptaConsentimiento: true, VersionTerminos: "2026.1",
			TextosConsentimiento: []string{"Autorizo el uso de mis datos personales.", "Declaro que la información es verdadera."},
		})
	}

	// 1. Incompleta: sin cargo elegible ni secciones.
	_, err = guardar(p.ID)
	var incompleta *application.PostulacionIncompletaError
	if !errors.As(err, &incompleta) {
		t.Fatalf("esperado PostulacionIncompletaError, obtenido %v", err)
	}
	texto := strings.Join(incompleta.Faltantes, " | ")
	if !strings.Contains(texto, "cargo") || !strings.Contains(texto, "Información personal") || strings.Contains(texto, "Capacitación") {
		t.Fatalf("faltantes inesperados: %s", texto)
	}

	// 2. Ajena: 403 sin revelar qué falta.
	if _, err := guardar("33333333-3333-3333-3333-333333333333"); !errors.Is(err, domain.ErrAccesoDenegado) {
		t.Fatalf("esperado ErrAccesoDenegado, obtenido %v", err)
	}

	// 3. Completa: cargo elegible y todas las secciones obligatorias completas (capacitación es opcional).
	if _, err := pool.Exec(ctx, `UPDATE postulacion SET cargo_id = 1, elegibilidad_estado = 'ELEGIBLE', elegibilidad_motivo = 'Postulante afiliado al IESS' WHERE id = $1::uuid`, borrador.ID); err != nil {
		t.Fatal(err)
	}
	for _, seccion := range domain.SeccionesFormulario {
		if seccion == "capacitacion" {
			continue
		}
		if _, err := pool.Exec(ctx, `INSERT INTO postulacion_seccion (postulacion_id, seccion, datos, completa) VALUES ($1::uuid, $2, '{}'::jsonb, true)
			ON CONFLICT (postulacion_id, seccion) DO UPDATE SET completa = true`, borrador.ID, seccion); err != nil {
			t.Fatal(err)
		}
	}
	res, err := guardar(p.ID)
	if err != nil || res.Estado != string(domain.EstadoListaParaEnvio) {
		t.Fatalf("guardar: esperado LISTA_PARA_ENVIO, obtenido %+v %v", res, err)
	}
	if res, err := guardar(p.ID); err != nil || res.Estado != string(domain.EstadoListaParaEnvio) {
		t.Fatalf("repetir guardar debía ser idempotente: %+v %v", res, err)
	}
	// Los textos aceptados quedan guardados y son los que imprime el formulario.
	textos, err := formularioRepo.TextosConsentimiento(ctx, borrador.ID)
	if err != nil || len(textos) != 2 || textos[1] != "Declaro que la información es verdadera." {
		t.Fatalf("textos de consentimiento: %v %v", textos, err)
	}
	if !strings.Contains(application.TextoConsentimientos(textos), "2. Declaro que la información es verdadera.") {
		t.Fatal("el texto del PDF debía numerar los consentimientos")
	}

	// 3b. Una versión de términos más larga que la columna responde un error de validación, no 500.
	if _, err := svc.GuardarParaEnvio(ctx, application.GuardarParaEnvioRequest{
		PostulacionID: borrador.ID, PostulanteID: p.ID, AceptaConsentimiento: true,
		VersionTerminos: "[PENDIENTE: versión de términos]",
	}); !errors.Is(err, application.ErrVersionTerminosInvalida) {
		t.Fatalf("versión larga: esperado ErrVersionTerminosInvalida, obtenido %v", err)
	}

	// 4. Ahora el envío funciona.
	enviada, err := svc.EnviarPostulacion(ctx, application.EnviarPostulacionRequest{
		PostulacionID: borrador.ID, PostulanteID: p.ID, AceptaConsentimiento: true,
		VersionTerminos: "2026.1", IdempotencyKey: fmt.Sprintf("guardar-%d", sufijo),
	})
	if err != nil || enviada.Estado != string(domain.EstadoEnviada) || enviada.CodigoUnico == "" {
		t.Fatalf("enviar tras guardar: %+v %v", enviada, err)
	}
	if _, err := guardar(p.ID); !errors.Is(err, application.ErrPostulacionYaEnviada) {
		t.Fatalf("guardar una enviada: esperado ErrPostulacionYaEnviada, obtenido %v", err)
	}

	// 5. Mi postulación: código, fecha de envío e historial en orden cronológico inverso.
	actual, err := formularioRepo.ObtenerActual(ctx, p.ID)
	if err != nil || actual.CodigoUnico != enviada.CodigoUnico || actual.FechaEnvio == nil {
		t.Fatalf("estado sin código o fecha de envío: %+v %v", actual, err)
	}
	historial, err := svc.Historial(ctx, borrador.ID, p.ID)
	if err != nil {
		t.Fatalf("Historial: %v", err)
	}
	estados := make([]string, 0, len(historial))
	for _, e := range historial {
		estados = append(estados, e.Estado)
	}
	if strings.Join(estados, ",") != "ENVIADA,LISTA_PARA_ENVIO,BORRADOR" {
		t.Fatalf("historial inesperado: %v", estados)
	}
	if _, err := svc.Historial(ctx, borrador.ID, "33333333-3333-3333-3333-333333333333"); !errors.Is(err, domain.ErrAccesoDenegado) {
		t.Fatalf("historial ajeno: esperado ErrAccesoDenegado, obtenido %v", err)
	}
}
