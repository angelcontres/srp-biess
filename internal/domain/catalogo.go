package domain

import "context"

// Pais, Provincia, Canton representan la división política y catálogos.
type Pais struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Codigo string `json:"codigo"`
}

type Provincia struct {
	ID     string `json:"id"`
	PaisID string `json:"pais_id"`
	Nombre string `json:"nombre"`
	Codigo string `json:"codigo"`
}

type Canton struct {
	ID          string `json:"id"`
	ProvinciaID string `json:"provincia_id"`
	Nombre      string `json:"nombre"`
}

type Cargo struct {
	ID                 string  `json:"id"`
	Nombre             string  `json:"nombre"`
	Descripcion        string  `json:"descripcion"`
	ExperienciaMinAnos int     `json:"experiencia_min_anos"`
	Arancel            float64 `json:"arancel"`
	Activo             bool    `json:"activo"`
}

// CatalogoRepository define el acceso a catálogos con estrategia de cache.
type CatalogoRepository interface {
	GetPaises(ctx context.Context) ([]Pais, error)
	GetProvincias(ctx context.Context, paisID string) ([]Provincia, error)
	GetCantones(ctx context.Context, provinciaID string) ([]Canton, error)
	GetCargos(ctx context.Context) ([]Cargo, error)
	GetCargoByID(ctx context.Context, id string) (*Cargo, error)
}
