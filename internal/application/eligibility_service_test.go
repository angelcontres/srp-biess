package application_test

import (
	"context"
	"errors"
	"testing"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/external"
)

func TestEligibilityService(t *testing.T) {
	svc := application.NewEligibilityService(external.NewMockEmploymentProvider())
	ctx := context.Background()

	tests := []struct {
		name     string
		cedula   string
		sector   domain.SectorCargo
		wantCode string // vacío = elegible
	}{
		{"Afiliado elegible", "0926687856", domain.SectorAfiliados, ""},
		{"Jubilado elegible", "1710034065", domain.SectorJubilados, ""},
		{"Empleador elegible", "0912345675", domain.SectorEmpleadores, ""},
		{"No afiliado", "1709876542", domain.SectorAfiliados, "POS_NO_ELEGIBLE_AFI"},
		{"No jubilado", "0926687856", domain.SectorJubilados, "POS_NO_ELEGIBLE_JUB"},
		{"No empleador", "1710034065", domain.SectorEmpleadores, "POS_NO_ELEGIBLE_EMP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Validate(ctx, tt.cedula, tt.sector)
			if tt.wantCode == "" {
				if err != nil {
					t.Fatalf("se esperaba elegible: %v", err)
				}
				return
			}
			var eligErr *domain.EligibilityError
			if !errors.As(err, &eligErr) {
				t.Fatalf("err = %v; want *EligibilityError", err)
			}
			if eligErr.Code != tt.wantCode {
				t.Errorf("Code = %s; want %s", eligErr.Code, tt.wantCode)
			}
			if !errors.Is(err, domain.ErrIneligibleCandidate) {
				t.Errorf("debe envolver ErrIneligibleCandidate")
			}
		})
	}
}
