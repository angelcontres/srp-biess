package external_test

import (
	"context"
	"errors"
	"testing"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/external"
)

func TestMockIdentityProvider(t *testing.T) {
	rc := external.NewMockIdentityProvider()
	ctx := context.Background()

	tests := []struct {
		name      string
		cedula    string
		tipo      domain.TipoValidacion
		valor     string
		wantMatch bool
		wantErr   error
	}{
		{"Fecha de expedición correcta", "0926687856", domain.ValidacionFechaExpedicion, "2019-05-14", true, nil},
		{"Código dactilar correcto", "1710034065", domain.ValidacionCodigoDactilar, "E3331I2222", true, nil},
		{"Fecha de expedición errónea", "0926687856", domain.ValidacionFechaExpedicion, "2000-01-01", false, nil},
		{"Código dactilar erróneo", "1710034065", domain.ValidacionCodigoDactilar, "X0000X0000", false, nil},
		{"Ciudadano inexistente", "1709876542", domain.ValidacionFechaExpedicion, "2019-05-14", false, nil},
		{"Tipo de validación inválido", "0926687856", "HUELLA", "x", false, domain.ErrInvalidTipoValidacion},
		{"Registro Civil caído", external.CedulaServicioCaido, domain.ValidacionFechaExpedicion, "x", false, domain.ErrExternalUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := rc.Validate(ctx, tt.cedula, tt.tipo, tt.valor)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v; want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if res.Match != tt.wantMatch {
				t.Errorf("Match = %v; want %v", res.Match, tt.wantMatch)
			}
			if !res.Match && res.Nombres != "" {
				t.Errorf("no se deben revelar datos biográficos sin coincidencia")
			}
		})
	}
}

func TestMockIdentityProviderNacionalidad(t *testing.T) {
	rc := external.NewMockIdentityProvider()
	res, err := rc.Validate(context.Background(), "1304567892", domain.ValidacionFechaExpedicion, "2021-08-17")
	if err != nil || !res.Match {
		t.Fatalf("se esperaba coincidencia: %v", err)
	}
	if res.EsEcuatoriano() {
		t.Errorf("ciudadano extranjero no debe ser ecuatoriano")
	}
}

func TestMockEmploymentProvider(t *testing.T) {
	iess := external.NewMockEmploymentProvider()
	ctx := context.Background()

	tests := []struct {
		name   string
		cedula string
		sector domain.SectorCargo
		cumple bool
	}{
		{"Afiliado en sector afiliados", "0926687856", domain.SectorAfiliados, true},
		{"Afiliado en sector jubilados", "0926687856", domain.SectorJubilados, false},
		{"Jubilado en sector jubilados", "1710034065", domain.SectorJubilados, true},
		{"Empleador en sector empleadores", "0912345675", domain.SectorEmpleadores, true},
		{"Jubilado en sector empleadores", "1710034065", domain.SectorEmpleadores, false},
		{"Sin registros IESS", "1709876542", domain.SectorAfiliados, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := iess.GetQuality(ctx, tt.cedula, tt.sector)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if got := q.CumpleSector(tt.sector); got != tt.cumple {
				t.Errorf("CumpleSector = %v; want %v", got, tt.cumple)
			}
		})
	}

	if _, err := iess.GetQuality(ctx, "0926687856", "OTRO"); !errors.Is(err, domain.ErrInvalidSector) {
		t.Errorf("sector inválido: err = %v", err)
	}
	if _, err := iess.GetQuality(ctx, external.CedulaServicioCaido, domain.SectorAfiliados); !errors.Is(err, domain.ErrExternalUnavailable) {
		t.Errorf("IESS caído: err = %v", err)
	}
}
