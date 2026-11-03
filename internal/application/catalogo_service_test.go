package application_test

import (
	"context"
	"errors"
	"testing"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
)

type catalogoRepoStub struct {
	paises []domain.Pais
	calls  int
}

func (r *catalogoRepoStub) GetPaises(context.Context) ([]domain.Pais, error) {
	r.calls++
	return r.paises, nil
}
func (r *catalogoRepoStub) GetProvincias(context.Context, string) ([]domain.Provincia, error) {
	return nil, nil
}
func (r *catalogoRepoStub) GetCantones(context.Context, string) ([]domain.Canton, error) {
	return nil, nil
}
func (r *catalogoRepoStub) GetCargos(context.Context) ([]domain.Cargo, error) { return nil, nil }
func (r *catalogoRepoStub) GetCargoByID(context.Context, string) (*domain.Cargo, error) {
	return nil, nil
}

type catalogoCacheStub struct {
	items []domain.Pais
	hit   bool
	sets  int
}

func (c *catalogoCacheStub) GetPaises(context.Context) ([]domain.Pais, bool, error) {
	return c.items, c.hit, nil
}
func (c *catalogoCacheStub) SetPaises(_ context.Context, items []domain.Pais) error {
	c.items, c.sets = items, c.sets+1
	return nil
}
func (c *catalogoCacheStub) GetProvincias(context.Context, string) ([]domain.Provincia, bool, error) {
	return nil, false, nil
}
func (c *catalogoCacheStub) SetProvincias(context.Context, string, []domain.Provincia) error {
	return nil
}
func (c *catalogoCacheStub) GetCantones(context.Context, string) ([]domain.Canton, bool, error) {
	return nil, false, nil
}
func (c *catalogoCacheStub) SetCantones(context.Context, string, []domain.Canton) error { return nil }
func (c *catalogoCacheStub) Invalidate(context.Context) error                           { return nil }

func TestCatalogoService_CacheAside(t *testing.T) {
	repo := &catalogoRepoStub{paises: []domain.Pais{{ID: "1", Nombre: "Ecuador"}}}
	cache := &catalogoCacheStub{}
	service := application.NewCatalogoService(repo, cache)

	items, err := service.ListarPaises(context.Background())
	if err != nil || len(items) != 1 || repo.calls != 1 || cache.sets != 1 {
		t.Fatalf("miss inesperado: items=%v err=%v repo=%d sets=%d", items, err, repo.calls, cache.sets)
	}

	cache.hit = true
	items, err = service.ListarPaises(context.Background())
	if err != nil || len(items) != 1 || repo.calls != 1 {
		t.Fatalf("hit no evitó PostgreSQL: items=%v err=%v repo=%d", items, err, repo.calls)
	}
}

func TestCatalogoService_FallbackWhenCacheFails(t *testing.T) {
	repo := &catalogoRepoStub{paises: []domain.Pais{{ID: "1"}}}
	cache := &failingCatalogoCache{}
	service := application.NewCatalogoService(repo, cache)

	items, err := service.ListarPaises(context.Background())
	if err != nil || len(items) != 1 || repo.calls != 1 {
		t.Fatalf("fallback inesperado: items=%v err=%v repo=%d", items, err, repo.calls)
	}
}

type failingCatalogoCache struct{}

func (failingCatalogoCache) GetPaises(context.Context) ([]domain.Pais, bool, error) {
	return nil, false, errors.New("redis unavailable")
}
func (failingCatalogoCache) SetPaises(context.Context, []domain.Pais) error {
	return errors.New("redis unavailable")
}
func (failingCatalogoCache) GetProvincias(context.Context, string) ([]domain.Provincia, bool, error) {
	return nil, false, nil
}
func (failingCatalogoCache) SetProvincias(context.Context, string, []domain.Provincia) error {
	return nil
}
func (failingCatalogoCache) GetCantones(context.Context, string) ([]domain.Canton, bool, error) {
	return nil, false, nil
}
func (failingCatalogoCache) SetCantones(context.Context, string, []domain.Canton) error { return nil }
func (failingCatalogoCache) Invalidate(context.Context) error                           { return nil }
