package database_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
)

// Prueba de integración del formulario (SRP-013). Requiere PostgreSQL y la migración
// db/migrations/001_postulacion_seccion.sql; si falta alguno, se omite.
func TestFormularioRepository_Integracion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	defer pool.Close()

	var existe bool
	_ = pool.QueryRow(ctx, `SELECT to_regclass('public.postulacion_seccion') IS NOT NULL`).Scan(&existe)
	if !existe {
		t.Skip("Falta la tabla postulacion_seccion: ejecute db/migrations/001_postulacion_seccion.sql")
	}

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
