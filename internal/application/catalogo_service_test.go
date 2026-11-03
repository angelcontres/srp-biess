package application_test

import (
	"context"
	"testing"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
)

// catalogoRepoFake devuelve un elemento por consulta y guarda el filtro recibido.
type catalogoRepoFake struct {
	filtro string
}

func (r *catalogoRepoFake) GetPaises(context.Context) ([]domain.Pais, error) {
	return []domain.Pais{{ID: "52", Nombre: "Ecuador", Codigo: "EC"}}, nil
}

func (r *catalogoRepoFake) GetProvincias(_ context.Context, paisID string) ([]domain.Provincia, error) {
	r.filtro = paisID
	return []domain.Provincia{{ID: "9", PaisID: paisID, Nombre: "Guayas"}}, nil
}

func (r *catalogoRepoFake) GetCantones(_ context.Context, provinciaID string) ([]domain.Canton, error) {
	r.filtro = provinciaID
	return []domain.Canton{{ID: "1", ProvinciaID: provinciaID, Nombre: "Guayaquil"}}, nil
}

func (r *catalogoRepoFake) GetCargos(context.Context) ([]domain.Cargo, error) {
	return []domain.Cargo{{}}, nil
}

func (r *catalogoRepoFake) GetCargoByID(context.Context, string) (*domain.Cargo, error) {
	return nil, domain.ErrNotFound
}

func TestCatalogoService_Delega(t *testing.T) {
	ctx := context.Background()
	repo := &catalogoRepoFake{}
	svc := application.NewCatalogoService(repo)

	if paises, err := svc.ListarPaises(ctx); err != nil || len(paises) != 1 || paises[0].Codigo != "EC" {
		t.Errorf("ListarPaises: %v %v", paises, err)
	}
	if provincias, err := svc.ListarProvincias(ctx, "52"); err != nil || len(provincias) != 1 || repo.filtro != "52" {
		t.Errorf("ListarProvincias: %v %v filtro=%q", provincias, err, repo.filtro)
	}
	if cantones, err := svc.ListarCantones(ctx, "9"); err != nil || len(cantones) != 1 || repo.filtro != "9" {
		t.Errorf("ListarCantones: %v %v filtro=%q", cantones, err, repo.filtro)
	}
	if cargos, err := svc.ListarCargos(ctx); err != nil || len(cargos) != 1 {
		t.Errorf("ListarCargos: %v %v", cargos, err)
	}
}
