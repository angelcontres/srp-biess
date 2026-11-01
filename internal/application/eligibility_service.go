package application

import (
	"context"

	"srp-biess/internal/domain"
)

// EligibilityService valida la calidad de afiliado, jubilado o empleador (SDD 5.1).
type EligibilityService struct {
	iess domain.EmploymentHistoryProvider
}

func NewEligibilityService(iess domain.EmploymentHistoryProvider) *EligibilityService {
	return &EligibilityService{iess: iess}
}

// Validate retorna *domain.EligibilityError con el mensaje reglamentario si no es elegible.
func (s *EligibilityService) Validate(ctx context.Context, cedula string, sector domain.SectorCargo) (*domain.EmploymentQuality, error) {
	if !sector.EsValido() {
		return nil, domain.ErrInvalidSector
	}
	q, err := s.iess.GetQuality(ctx, cedula, sector)
	if err != nil {
		return nil, err
	}
	if !q.CumpleSector(sector) {
		return q, domain.NewEligibilityError(sector)
	}
	return q, nil
}
