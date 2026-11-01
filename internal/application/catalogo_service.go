package application

import (
	"context"
	"srp-biess/internal/domain"
)

type CatalogoService struct {
	repo domain.CatalogoRepository
}

func NewCatalogoService(repo domain.CatalogoRepository) *CatalogoService {
	return &CatalogoService{repo: repo}
}

func (s *CatalogoService) ListarPaises(ctx context.Context) ([]domain.Pais, error) {
	return s.repo.GetPaises(ctx)
}

func (s *CatalogoService) ListarProvincias(ctx context.Context, paisID string) ([]domain.Provincia, error) {
	return s.repo.GetProvincias(ctx, paisID)
}

func (s *CatalogoService) ListarCantones(ctx context.Context, provinciaID string) ([]domain.Canton, error) {
	return s.repo.GetCantones(ctx, provinciaID)
}

func (s *CatalogoService) ListarCargos(ctx context.Context) ([]domain.Cargo, error) {
	return s.repo.GetCargos(ctx)
}
