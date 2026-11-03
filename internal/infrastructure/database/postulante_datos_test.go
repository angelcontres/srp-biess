package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
)

// Los datos del Registro Civil se guardan al registrar y llegan al formulario PDF (RF-EF12-03).
func TestPostulante_GuardaDatosRegistroCivil(t *testing.T) {
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

	sufijo := time.Now().UnixNano()
	nacimiento := time.Date(1985, 3, 21, 0, 0, 0, 0, time.UTC)
	p := &domain.Postulante{
		Cedula: fmt.Sprintf("03%08d", sufijo%100000000), Nombres: "MARIA JOSE", Apellidos: "PEREZ LOPEZ",
		Email: fmt.Sprintf("rc_%d@biess.gob.ec", sufijo), PasswordHash: "hash_seguro",
		Genero: "FEMENINO", FechaNacimiento: &nacimiento, EstadoCivil: "CASADO",
	}
	if err := database.NewPostgresPostulanteRepository(pool).Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Cleanup(func() {
		bg := context.Background()
		_, _ = pool.Exec(bg, `DELETE FROM postulacion WHERE postulante_id = $1::uuid`, p.ID)
		_, _ = pool.Exec(bg, `DELETE FROM postulante WHERE id = $1::uuid`, p.ID)
	})

	var genero, fecha, civil, nacionalidad string
	if err := pool.QueryRow(ctx, `SELECT genero::text, fecha_nacimiento::text, estado_civil::text, nacionalidad FROM postulante WHERE id = $1::uuid`, p.ID).
		Scan(&genero, &fecha, &civil, &nacionalidad); err != nil {
		t.Fatal(err)
	}
	if genero != "FEMENINO" || fecha != "1985-03-21" || civil != "CASADO" || nacionalidad != "ECUATORIANA" {
		t.Fatalf("datos guardados inesperados: %s %s %s %s", genero, fecha, civil, nacionalidad)
	}

	// Llegan a los datos verificados y al formulario PDF.
	formulario := database.NewPostgresFormularioRepository(pool)
	borrador, err := formulario.CrearBorrador(ctx, p.ID)
	if err != nil {
		t.Skipf("sin convocatoria vigente para comprobar el detalle: %v", err)
	}
	detalle, err := formulario.ObtenerDetalle(ctx, borrador.ID, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	f := application.ArmarFormularioPDF(detalle, "X", time.Now(), "", "", nil)
	valores := map[string]string{}
	for _, c := range f.Identificacion {
		valores[c.Etiqueta] = c.Valor
	}
	if valores["Género"] != "Femenino" || valores["Fecha de nacimiento"] != "21/03/1985" {
		t.Fatalf("identificación del PDF incompleta: %+v", valores)
	}
}
