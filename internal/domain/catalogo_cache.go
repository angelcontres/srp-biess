package domain

import "context"

// CatalogoCache define el puerto distribuido para catálogos cacheados.
type CatalogoCache interface {
	GetPaises(ctx context.Context) ([]Pais, bool, error)
	SetPaises(ctx context.Context, paises []Pais) error
	GetProvincias(ctx context.Context, paisID string) ([]Provincia, bool, error)
	SetProvincias(ctx context.Context, paisID string, provincias []Provincia) error
	GetCantones(ctx context.Context, provinciaID string) ([]Canton, bool, error)
	SetCantones(ctx context.Context, provinciaID string, cantones []Canton) error
	Invalidate(ctx context.Context) error
}
