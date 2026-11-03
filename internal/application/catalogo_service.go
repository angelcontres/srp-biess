package application

import (
	"context"

	"srp-biess/internal/domain"
)

type CatalogoService struct {
	repo  domain.CatalogoRepository
	cache domain.CatalogoCache
}

func NewCatalogoService(repo domain.CatalogoRepository, cache ...domain.CatalogoCache) *CatalogoService {
	var catalogoCache domain.CatalogoCache
	if len(cache) > 0 {
		catalogoCache = cache[0]
	}
	return &CatalogoService{repo: repo, cache: catalogoCache}
}

func (s *CatalogoService) ListarPaises(ctx context.Context) ([]domain.Pais, error) {
	if s.cache != nil {
		if paises, ok, err := s.cache.GetPaises(ctx); err == nil && ok {
			return paises, nil
		}
	}
	paises, err := s.repo.GetPaises(ctx)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetPaises(ctx, paises)
	}
	return paises, nil
}

func (s *CatalogoService) ListarProvincias(ctx context.Context, paisID string) ([]domain.Provincia, error) {
	if s.cache != nil {
		if provincias, ok, err := s.cache.GetProvincias(ctx, paisID); err == nil && ok {
			return provincias, nil
		}
	}
	provincias, err := s.repo.GetProvincias(ctx, paisID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetProvincias(ctx, paisID, provincias)
	}
	return provincias, nil
}

func (s *CatalogoService) ListarCantones(ctx context.Context, provinciaID string) ([]domain.Canton, error) {
	if s.cache != nil {
		if cantones, ok, err := s.cache.GetCantones(ctx, provinciaID); err == nil && ok {
			return cantones, nil
		}
	}
	cantones, err := s.repo.GetCantones(ctx, provinciaID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetCantones(ctx, provinciaID, cantones)
	}
	return cantones, nil
}

func (s *CatalogoService) ListarCargos(ctx context.Context) ([]domain.Cargo, error) {
	return s.repo.GetCargos(ctx)
}
