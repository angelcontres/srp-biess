package external

import (
	"context"

	"srp-biess/internal/domain"
)

// MockEmploymentProvider simula el Historial Laboral del IESS con datos deterministas (ADR-010).
type MockEmploymentProvider struct {
	calidades map[string]domain.EmploymentQuality
}

func NewMockEmploymentProvider() *MockEmploymentProvider {
	return &MockEmploymentProvider{calidades: map[string]domain.EmploymentQuality{
		"0926687856": {Cedula: "0926687856", EsAfiliado: true},
		"1710034065": {Cedula: "1710034065", EsJubilado: true},
		"0912345675": {Cedula: "0912345675", EsAfiliado: true, EsEmpleador: true},
	}}
}

func (m *MockEmploymentProvider) GetQuality(ctx context.Context, cedula string, sector domain.SectorCargo) (*domain.EmploymentQuality, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !sector.EsValido() {
		return nil, domain.ErrInvalidSector
	}
	if cedula == CedulaServicioCaido {
		return nil, domain.ErrExternalUnavailable
	}

	q, ok := m.calidades[cedula]
	if !ok {
		// Ciudadano sin registros en el IESS: no cumple ninguna calidad.
		return &domain.EmploymentQuality{Cedula: cedula}, nil
	}
	return &q, nil
}
