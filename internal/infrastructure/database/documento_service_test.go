package database_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
	"srp-biess/internal/infrastructure/security"
	"srp-biess/internal/infrastructure/storage"
)

func pdf(tamano int) []byte {
	b := bytes.Repeat([]byte("x"), tamano)
	copy(b, "%PDF-1.7\n")
	return b
}

type entornoDocumentos struct {
	pool    *pgxpool.Pool
	svc     *application.DocumentoService
	raiz    string
	limites domain.LimitesDocumento
}

// nuevoBorradorDocumentos crea un postulante con su borrador (requiere una convocatoria vigente).
func nuevoBorradorDocumentos(ctx context.Context, t *testing.T, pool *pgxpool.Pool, etiqueta string) (postulanteID, postulacionID string) {
	t.Helper()
	sufijo := time.Now().UnixNano()
	p := &domain.Postulante{
		Cedula: fmt.Sprintf("07%08d", sufijo%100000000), Nombres: "PRUEBA", Apellidos: "DOCUMENTOS",
		Email: fmt.Sprintf("docs_%s_%d@biess.gob.ec", etiqueta, sufijo), PasswordHash: "hash_seguro",
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
	b, err := database.NewPostgresFormularioRepository(pool).CrearBorrador(ctx, p.ID)
	if errors.Is(err, domain.ErrFaseCerrada) {
		t.Skip("No hay convocatoria con la fase habilitada en la fecha actual")
	}
	if err != nil {
		t.Fatalf("CrearBorrador: %v", err)
	}
	return p.ID, b.ID
}

func (e entornoDocumentos) filas(ctx context.Context, t *testing.T, postulacionID string) int {
	t.Helper()
	var n int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM documento WHERE postulacion_id = $1::uuid`, postulacionID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (e entornoDocumentos) archivos(t *testing.T) (definitivos, temporales int) {
	t.Helper()
	ents, _ := os.ReadDir(filepath.Join(e.raiz, "documentos"))
	for _, ent := range ents {
		if !ent.IsDir() {
			definitivos++
		}
	}
	tmp, _ := os.ReadDir(filepath.Join(e.raiz, "documentos", "tmp"))
	return definitivos, len(tmp)
}

// Prueba de integración de SRP-017b contra PostgreSQL y el sistema de archivos.
func TestDocumentoService_Integracion(t *testing.T) {
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

	// Límites pequeños para probar cuotas sin archivos grandes; respaldo hasta 4 MB como en la base.
	limites := domain.LimitesDocumento{RespaldoBytes: 4 * 1024 * 1024, OtrosBytes: 2000, TotalBytes: 5000}
	nuevoEntorno := func(t *testing.T) entornoDocumentos {
		raiz := t.TempDir()
		almacen, err := storage.NewAlmacenLocal(raiz)
		if err != nil {
			t.Fatal(err)
		}
		svc := application.NewDocumentoService(database.NewPgxTxManager(pool), database.NewPostgresDocumentoRepository(pool),
			almacen, security.NewFirmadorHMAC([]byte("secreto-integracion")), database.NewPostgresAuditoriaRepository(pool),
			limites, time.Minute)
		return entornoDocumentos{pool: pool, svc: svc, raiz: raiz, limites: limites}
	}
	sol := domain.ContextoSolicitud{IP: "192.0.2.20", UserAgent: "prueba-documentos"}

	t.Run("un PDF válido queda con hash, UUID, archivo y auditoría; luego se descarga y se elimina", func(t *testing.T) {
		e := nuevoEntorno(t)
		postulante, postulacion := nuevoBorradorDocumentos(ctx, t, pool, "valido")
		contenido := pdf(1500)

		cargado, err := e.svc.Subir(ctx, sol, postulacion, postulante, domain.DocTitulo, "../Título Ingeniería.pdf", bytes.NewReader(contenido))
		if err != nil {
			t.Fatalf("Subir: %v", err)
		}
		suma := sha256.Sum256(contenido)
		d := cargado.Documento
		if d.HashSHA256 != hex.EncodeToString(suma[:]) || d.TamanoBytes != 1500 || d.NombreOriginal != "Título Ingeniería.pdf" ||
			len(d.NombreAlmacenado) != 40 || d.CreatedAt.IsZero() {
			t.Fatalf("documento inesperado: %+v", d)
		}
		if def, tmp := e.archivos(t); def != 1 || tmp != 0 || e.filas(ctx, t, postulacion) != 1 {
			t.Fatalf("esperado 1 archivo y 1 fila, sin temporales: %d %d", def, tmp)
		}
		var accion string
		if err := pool.QueryRow(ctx, `SELECT accion FROM auditoria_evento WHERE entidad = 'documento' AND entidad_id = $1`, d.ID).Scan(&accion); err != nil || accion != domain.AccionDocumentoCargado {
			t.Fatalf("la carga debía auditarse: %s %v", accion, err)
		}

		// Descarga con la firma de la URL devuelta.
		u, _ := url.Parse(cargado.URLDescarga)
		vence, _ := strconv.ParseInt(u.Query().Get("exp"), 10, 64)
		_, archivo, err := e.svc.Descargar(ctx, postulacion, postulante, d.ID, vence, u.Query().Get("sig"))
		if err != nil {
			t.Fatalf("Descargar: %v", err)
		}
		leido, _ := io.ReadAll(archivo)
		_ = archivo.Close()
		if !bytes.Equal(leido, contenido) {
			t.Fatal("el contenido descargado no coincide")
		}
		if _, _, err := e.svc.Descargar(ctx, postulacion, postulante, d.ID, vence, "00"+u.Query().Get("sig")[2:]); !errors.Is(err, domain.ErrFirmaDescargaInvalida) {
			t.Fatalf("firma alterada: esperado ErrFirmaDescargaInvalida, obtenido %v", err)
		}
		vencido := e.svc.WithReloj(func() time.Time { return time.Now().Add(2 * time.Minute) })
		if _, _, err := vencido.Descargar(ctx, postulacion, postulante, d.ID, vence, u.Query().Get("sig")); !errors.Is(err, domain.ErrFirmaDescargaInvalida) {
			t.Fatalf("firma vencida: esperado ErrFirmaDescargaInvalida, obtenido %v", err)
		}
		vencido.WithReloj(time.Now)

		// Otro postulante no ve el documento, ni con firma válida.
		otro, otraPostulacion := nuevoBorradorDocumentos(ctx, t, pool, "otro")
		if _, _, err := e.svc.Descargar(ctx, postulacion, otro, d.ID, vence, u.Query().Get("sig")); !errors.Is(err, domain.ErrDocumentoNoEncontrado) {
			t.Fatalf("documento ajeno: esperado ErrDocumentoNoEncontrado, obtenido %v", err)
		}
		if err := e.svc.Eliminar(ctx, sol, otraPostulacion, otro, d.ID); !errors.Is(err, domain.ErrDocumentoNoEncontrado) {
			t.Fatalf("eliminar desde otra postulación: esperado ErrDocumentoNoEncontrado, obtenido %v", err)
		}

		if err := e.svc.Eliminar(ctx, sol, postulacion, postulante, d.ID); err != nil {
			t.Fatalf("Eliminar: %v", err)
		}
		if def, _ := e.archivos(t); def != 0 || e.filas(ctx, t, postulacion) != 0 {
			t.Fatalf("eliminar debía borrar fila y archivo: %d archivos", def)
		}
		if err := pool.QueryRow(ctx, `SELECT accion FROM auditoria_evento WHERE entidad = 'documento' AND entidad_id = $1 AND accion = $2`, d.ID, domain.AccionDocumentoEliminado).Scan(&accion); err != nil {
			t.Fatalf("el borrado debía auditarse: %v", err)
		}
	})

	t.Run("rechazos sin dejar archivos ni filas", func(t *testing.T) {
		e := nuevoEntorno(t)
		postulante, postulacion := nuevoBorradorDocumentos(ctx, t, pool, "rechazos")
		casos := []struct {
			nombre    string
			tipo      domain.TipoDocumento
			archivo   string
			contenido []byte
			esperado  error
		}{
			{"JPG renombrado", domain.DocTitulo, "cedula.pdf", []byte("\xFF\xD8\xFF\xE0 jpeg"), domain.ErrDocumentoNoPDF},
			{"extensión no PDF", domain.DocTitulo, "titulo.exe", pdf(100), domain.ErrDocumentoNoPDF},
			{"vacío", domain.DocTitulo, "vacio.pdf", nil, domain.ErrDocumentoNoPDF},
			{"excede 2 KB de prueba", domain.DocTitulo, "grande.pdf", pdf(2001), domain.ErrDocumentoExcedeTamano},
			{"FORMULARIO_POSTULACION", domain.DocFormularioPostulacion, "f.pdf", pdf(100), domain.ErrDocumentoTipoNoCargable},
			{"tipo desconocido", domain.TipoDocumento("OTRO"), "f.pdf", pdf(100), domain.ErrDocumentoTipoNoCargable},
		}
		for _, c := range casos {
			if _, err := e.svc.Subir(ctx, sol, postulacion, postulante, c.tipo, c.archivo, bytes.NewReader(c.contenido)); !errors.Is(err, c.esperado) {
				t.Errorf("%s: esperado %v, obtenido %v", c.nombre, c.esperado, err)
			}
		}
		if _, err := e.svc.Subir(ctx, sol, postulacion, "33333333-3333-3333-3333-333333333333", domain.DocTitulo, "a.pdf", bytes.NewReader(pdf(100))); !errors.Is(err, domain.ErrAccesoDenegado) {
			t.Errorf("postulación ajena: esperado ErrAccesoDenegado, obtenido %v", err)
		}
		if _, err := e.svc.Subir(ctx, sol, "44444444-4444-4444-4444-444444444444", postulante, domain.DocTitulo, "a.pdf", bytes.NewReader(pdf(100))); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("postulación inexistente: esperado ErrNotFound, obtenido %v", err)
		}
		if def, tmp := e.archivos(t); def != 0 || tmp != 0 || e.filas(ctx, t, postulacion) != 0 {
			t.Fatalf("los rechazos no debían dejar archivos (%d, %d temporales) ni filas", def, tmp)
		}
	})

	t.Run("el CHECK de la base se traduce a ErrDocumentoExcedeTamano", func(t *testing.T) {
		// Límites por encima de 4 MB: solo el CHECK ck_documento_tamano lo detiene.
		raiz := t.TempDir()
		almacen, _ := storage.NewAlmacenLocal(raiz)
		svc := application.NewDocumentoService(database.NewPgxTxManager(pool), database.NewPostgresDocumentoRepository(pool),
			almacen, security.NewFirmadorHMAC([]byte("s")), database.NewPostgresAuditoriaRepository(pool),
			domain.LimitesDocumento{RespaldoBytes: 5 * 1024 * 1024, OtrosBytes: 5 * 1024 * 1024, TotalBytes: 50 * 1024 * 1024}, time.Minute)
		postulante, postulacion := nuevoBorradorDocumentos(ctx, t, pool, "check")
		if _, err := svc.Subir(ctx, sol, postulacion, postulante, domain.DocRespaldoOrganizacional, "r.pdf", bytes.NewReader(pdf(4*1024*1024+1))); !errors.Is(err, domain.ErrDocumentoExcedeTamano) {
			t.Fatalf("esperado ErrDocumentoExcedeTamano por el CHECK, obtenido %v", err)
		}
		e := entornoDocumentos{pool: pool, raiz: raiz}
		if def, tmp := e.archivos(t); def != 0 || tmp != 0 || e.filas(ctx, t, postulacion) != 0 {
			t.Fatalf("el rechazo del CHECK no debía dejar archivos (%d, %d) ni filas", def, tmp)
		}
	})

	t.Run("cuota acumulada con cargas simultáneas: solo entra la que cabe", func(t *testing.T) {
		e := nuevoEntorno(t)
		postulante, postulacion := nuevoBorradorDocumentos(ctx, t, pool, "cuota")
		// 2 x 1900 = 3800 de 5000; quedan 1200: de dos cargas simultáneas de 1100 solo cabe una.
		for i := 0; i < 2; i++ {
			if _, err := e.svc.Subir(ctx, sol, postulacion, postulante, domain.DocTitulo, "t.pdf", bytes.NewReader(pdf(1900))); err != nil {
				t.Fatalf("carga previa %d: %v", i, err)
			}
		}
		var wg sync.WaitGroup
		errs := make([]error, 2)
		for i := range errs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				_, errs[i] = e.svc.Subir(ctx, sol, postulacion, postulante, domain.DocExperiencia, "e.pdf", bytes.NewReader(pdf(1100)))
			}(i)
		}
		wg.Wait()
		ok, cuota := 0, 0
		for _, err := range errs {
			switch {
			case err == nil:
				ok++
			case errors.Is(err, domain.ErrDocumentoCuotaExcedida):
				cuota++
			default:
				t.Fatalf("error inesperado: %v", err)
			}
		}
		if ok != 1 || cuota != 1 || e.filas(ctx, t, postulacion) != 3 {
			t.Fatalf("esperada 1 carga aceptada y 1 por cuota: ok=%d cuota=%d filas=%d", ok, cuota, e.filas(ctx, t, postulacion))
		}
		if def, tmp := e.archivos(t); def != 3 || tmp != 0 {
			t.Fatalf("esperados 3 archivos sin temporales: %d %d", def, tmp)
		}
	})

	t.Run("una postulación fuera de BORRADOR no admite carga ni borrado", func(t *testing.T) {
		e := nuevoEntorno(t)
		postulante, postulacion := nuevoBorradorDocumentos(ctx, t, pool, "cerrada")
		cargado, err := e.svc.Subir(ctx, sol, postulacion, postulante, domain.DocTitulo, "t.pdf", bytes.NewReader(pdf(100)))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `UPDATE postulacion SET estado = 'LISTA_PARA_ENVIO' WHERE id = $1::uuid`, postulacion); err != nil {
			t.Fatal(err)
		}
		if _, err := e.svc.Subir(ctx, sol, postulacion, postulante, domain.DocTitulo, "t.pdf", bytes.NewReader(pdf(100))); !errors.Is(err, domain.ErrPostulacionCerrada) {
			t.Fatalf("carga: esperado ErrPostulacionCerrada, obtenido %v", err)
		}
		if err := e.svc.Eliminar(ctx, sol, postulacion, postulante, cargado.Documento.ID); !errors.Is(err, domain.ErrPostulacionCerrada) {
			t.Fatalf("borrado: esperado ErrPostulacionCerrada, obtenido %v", err)
		}
		if def, _ := e.archivos(t); def != 1 || e.filas(ctx, t, postulacion) != 1 {
			t.Fatal("el rechazo no debía tocar el documento existente")
		}
	})
}
