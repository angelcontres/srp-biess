package database_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
)

// aplicarMigraciones ejecuta db/migrations/*.sql en orden. Son idempotentes, por lo que la
// prueba funciona también con volúmenes de PostgreSQL creados antes de la migración
// (docker-entrypoint-initdb.d solo ejecuta db/init.sql sobre un volumen vacío).
func aplicarMigraciones(ctx context.Context, t *testing.T, exec func(context.Context, string) error) {
	t.Helper()
	_, archivo, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(archivo), "..", "..", "..", "db", "migrations")
	rutas, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil || len(rutas) == 0 {
		t.Fatalf("no se encontraron migraciones en %s: %v", dir, err)
	}
	sort.Strings(rutas)
	for _, ruta := range rutas {
		sql, err := os.ReadFile(ruta)
		if err != nil {
			t.Fatalf("leer %s: %v", ruta, err)
		}
		if err := exec(ctx, string(sql)); err != nil {
			t.Fatalf("aplicar %s: %v", filepath.Base(ruta), err)
		}
	}
}

// Prueba de integración del formulario (SRP-013). Requiere PostgreSQL; aplica las
// migraciones de db/migrations antes de ejecutarse.
func TestFormularioRepository_Integracion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	postulante := &domain.Postulante{
		Cedula:       fmt.Sprintf("09%08d", sufijo%100000000),
		Nombres:      "PRUEBA",
		Apellidos:    "FORMULARIO",
		Email:        fmt.Sprintf("formulario_%d@biess.gob.ec", sufijo),
		PasswordHash: "hash_seguro",
	}
	if err := database.NewPostgresPostulanteRepository(pool).Save(ctx, postulante); err != nil {
		t.Fatalf("no se pudo crear el postulante: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM postulacion WHERE postulante_id = $1::uuid`, postulante.ID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM postulante WHERE id = $1::uuid`, postulante.ID)
	})

	repo := database.NewPostgresFormularioRepository(pool)

	if _, err := repo.ObtenerActual(ctx, postulante.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("esperado ErrNotFound sin borrador, obtenido %v", err)
	}

	borrador, err := repo.CrearBorrador(ctx, postulante.ID)
	if errors.Is(err, domain.ErrFaseCerrada) {
		t.Skip("No hay convocatoria con la fase habilitada en la fecha actual")
	}
	if err != nil {
		t.Fatalf("CrearBorrador: %v", err)
	}
	if _, err := repo.CrearBorrador(ctx, postulante.ID); !errors.Is(err, domain.ErrPostulacionExists) {
		t.Fatalf("el segundo borrador debía fallar con ErrPostulacionExists, obtenido %v", err)
	}

	estado, err := repo.GuardarSeccion(ctx, borrador.ID, postulante.ID, "domicilio", []byte(`{"callePrincipal":"Av. Amazonas"}`), true, borrador.Version)
	if err != nil {
		t.Fatalf("GuardarSeccion: %v", err)
	}
	if estado.Version != borrador.Version+1 || estado.Secciones["domicilio"] != domain.SeccionCompleta || estado.SeccionReanudacion != "domicilio" {
		t.Fatalf("estado inesperado tras guardar: %+v", estado)
	}

	if _, err := repo.GuardarSeccion(ctx, borrador.ID, postulante.ID, "domicilio", []byte(`{}`), false, borrador.Version); !errors.Is(err, domain.ErrConflictoVersion) {
		t.Fatalf("una versión desactualizada debía fallar con ErrConflictoVersion, obtenido %v", err)
	}

	detalle, err := repo.ObtenerDetalle(ctx, borrador.ID, postulante.ID)
	if err != nil {
		t.Fatalf("ObtenerDetalle: %v", err)
	}
	if string(detalle.Secciones["domicilio"]) != `{"callePrincipal": "Av. Amazonas"}` || detalle.Verificados.Cedula != postulante.Cedula {
		t.Fatalf("detalle inesperado: %s %+v", detalle.Secciones["domicilio"], detalle.Verificados)
	}

	if _, err := repo.ObtenerDetalle(ctx, borrador.ID, "33333333-3333-3333-3333-333333333333"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("otro postulante no debía ver la postulación, obtenido %v", err)
	}
}
