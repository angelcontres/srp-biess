package application_test

import (
	"context"
	"errors"
	"testing"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
)

// formularioRepoFake registra qué operaciones llegaron al repositorio.
type formularioRepoFake struct {
	actual       *domain.EstadoFormulario
	errActual    error
	detalle      *domain.DetalleFormulario
	errDetalle   error
	creados      int
	guardados    int
	ultimoGuardo string
}

func (r *formularioRepoFake) ObtenerActual(context.Context, string) (*domain.EstadoFormulario, error) {
	return r.actual, r.errActual
}

func (r *formularioRepoFake) CrearBorrador(context.Context, string) (*domain.EstadoFormulario, error) {
	r.creados++
	return &domain.EstadoFormulario{ID: "nuevo", Estado: "BORRADOR", Version: 1}, nil
}

func (r *formularioRepoFake) ObtenerDetalle(context.Context, string, string) (*domain.DetalleFormulario, error) {
	if r.errDetalle != nil {
		return nil, r.errDetalle
	}
	if r.detalle != nil {
		return r.detalle, nil
	}
	return &domain.DetalleFormulario{
		Postulacion: &domain.EstadoFormulario{Estado: "BORRADOR"},
		Verificados: &domain.DatosVerificados{Cedula: "0926687856", Correo: "postulante@biess.gob.ec"},
	}, nil
}

func (r *formularioRepoFake) GuardarSeccion(_ context.Context, _, _, seccion string, _ []byte, _ bool, version int) (*domain.EstadoFormulario, error) {
	r.guardados++
	r.ultimoGuardo = seccion
	return &domain.EstadoFormulario{Version: version + 1}, nil
}

func TestFormularioService_SinPersistencia(t *testing.T) {
	svc := application.NewFormularioService(nil)
	ctx := context.Background()

	errores := map[string]error{}
	_, errores["ObtenerActual"] = svc.ObtenerActual(ctx, "p")
	_, errores["CrearBorrador"] = svc.CrearBorrador(ctx, "p")
	_, errores["ObtenerDetalle"] = svc.ObtenerDetalle(ctx, "id", "p")
	_, errores["GuardarSeccion"] = svc.GuardarSeccion(ctx, "id", "p", "domicilio", []byte(`{}`), true, 1)
	for operacion, err := range errores {
		if !errors.Is(err, application.ErrPersistenciaNoDisponible) {
			t.Errorf("%s sin PostgreSQL: esperado ErrPersistenciaNoDisponible, obtenido %v", operacion, err)
		}
	}
}

func TestFormularioService_CrearBorrador(t *testing.T) {
	ctx := context.Background()

	t.Run("crea si no hay postulación vigente", func(t *testing.T) {
		repo := &formularioRepoFake{errActual: domain.ErrNotFound}
		estado, err := application.NewFormularioService(repo).CrearBorrador(ctx, "p")
		if err != nil || estado.ID != "nuevo" || repo.creados != 1 {
			t.Fatalf("esperado borrador nuevo: estado=%+v err=%v creados=%d", estado, err, repo.creados)
		}
	})

	t.Run("RF-EF12-01: no duplica un borrador vigente", func(t *testing.T) {
		repo := &formularioRepoFake{actual: &domain.EstadoFormulario{ID: "existente"}}
		if _, err := application.NewFormularioService(repo).CrearBorrador(ctx, "p"); !errors.Is(err, domain.ErrPostulacionExists) {
			t.Fatalf("esperado ErrPostulacionExists, obtenido %v", err)
		}
		if repo.creados != 0 {
			t.Error("no debía crearse un segundo borrador")
		}
	})

	t.Run("propaga un error técnico sin crear", func(t *testing.T) {
		falla := errors.New("conexión perdida")
		repo := &formularioRepoFake{errActual: falla}
		if _, err := application.NewFormularioService(repo).CrearBorrador(ctx, "p"); !errors.Is(err, falla) {
			t.Fatalf("esperado el error del repositorio, obtenido %v", err)
		}
		if repo.creados != 0 {
			t.Error("con un error técnico no debía crearse el borrador")
		}
	})
}

func TestFormularioService_GuardarSeccion(t *testing.T) {
	ctx := context.Background()

	invalidos := []struct {
		nombre, seccion, datos string
	}{
		{"sección desconocida", "pagos", `{}`},
		{"resumen no es editable", "resumen", `{}`},
		{"JSON mal formado", "domicilio", `{"calle":`},
		{"arreglo en vez de objeto", "domicilio", `[1,2]`},
		{"texto en vez de objeto", "domicilio", `"hola"`},
	}
	for _, c := range invalidos {
		repo := &formularioRepoFake{}
		_, err := application.NewFormularioService(repo).GuardarSeccion(ctx, "id", "p", c.seccion, []byte(c.datos), true, 1)
		if !errors.Is(err, domain.ErrSeccionInvalida) {
			t.Errorf("%s: esperado ErrSeccionInvalida, obtenido %v", c.nombre, err)
		}
		if repo.guardados != 0 {
			t.Errorf("%s: no debía llegar al repositorio", c.nombre)
		}
	}

	t.Run("guardado en borrador permite datos incompletos", func(t *testing.T) {
		repo := &formularioRepoFake{}
		estado, err := application.NewFormularioService(repo).GuardarSeccion(ctx, "id", "p", "experiencia", []byte(`{"experiencias":[]}`), false, 4)
		if err != nil || estado.Version != 5 || repo.ultimoGuardo != "experiencia" {
			t.Fatalf("guardado válido: estado=%+v err=%v sección=%q", estado, err, repo.ultimoGuardo)
		}
	})

	t.Run("completa true con datos incompletos no llama al repositorio", func(t *testing.T) {
		repo := &formularioRepoFake{}
		_, err := application.NewFormularioService(repo).GuardarSeccion(ctx, "id", "p", "personal", []byte(`{}`), true, 1)
		if err == nil {
			t.Fatal("esperaba error de validación")
		}
		var valErr *domain.ErrorValidacionSeccion
		if !errors.As(err, &valErr) {
			t.Fatalf("esperado *domain.ErrorValidacionSeccion, obtenido %T: %v", err, err)
		}
		if repo.guardados != 0 {
			t.Error("no debió llamar a GuardarSeccion en el repositorio")
		}
	})

	t.Run("completa true con datos válidos llama al repositorio", func(t *testing.T) {
		repo := &formularioRepoFake{}
		datos := []byte(`{"estadoCivil": "SOLTERO"}`)
		estado, err := application.NewFormularioService(repo).GuardarSeccion(ctx, "id", "p", "personal", datos, true, 1)
		if err != nil || repo.guardados != 1 || estado.Version != 2 {
			t.Fatalf("esperado guardado exitoso: err=%v guardados=%d", err, repo.guardados)
		}
	})

	t.Run("postulación cerrada (no BORRADOR) responde ErrPostulacionCerrada", func(t *testing.T) {
		repo := &formularioRepoFake{
			detalle: &domain.DetalleFormulario{
				Postulacion: &domain.EstadoFormulario{Estado: "ENVIADA"},
			},
		}
		datos := []byte(`{"estadoCivil": "SOLTERO"}`)
		_, err := application.NewFormularioService(repo).GuardarSeccion(ctx, "id", "p", "personal", datos, true, 1)
		if !errors.Is(err, domain.ErrPostulacionCerrada) {
			t.Fatalf("esperado ErrPostulacionCerrada, obtenido %v", err)
		}
		if repo.guardados != 0 {
			t.Error("no debió persistirse nada si la postulación está cerrada")
		}
	})
}

func TestFormularioService_DelegaConsultas(t *testing.T) {
	ctx := context.Background()
	repo := &formularioRepoFake{actual: &domain.EstadoFormulario{ID: "vigente"}}
	svc := application.NewFormularioService(repo)

	if estado, err := svc.ObtenerActual(ctx, "p"); err != nil || estado.ID != "vigente" {
		t.Errorf("ObtenerActual: estado=%+v err=%v", estado, err)
	}
	if detalle, err := svc.ObtenerDetalle(ctx, "vigente", "p"); err != nil || detalle == nil {
		t.Errorf("ObtenerDetalle: detalle=%v err=%v", detalle, err)
	}
}
