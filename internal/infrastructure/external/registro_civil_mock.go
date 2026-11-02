package external

import (
	"context"
	"time"

	"srp-biess/internal/domain"
)

// CedulaServicioCaido simula una caída del Registro Civil (prueba de resiliencia).
const CedulaServicioCaido = "0920000007"

type ciudadanoMock struct {
	fechaExpedicion string // YYYY-MM-DD
	codigoDactilar  string
	datos           domain.IdentityResult
}

// MockIdentityProvider simula el Registro Civil con datos deterministas (ADR-010).
type MockIdentityProvider struct {
	ciudadanos map[string]ciudadanoMock
}

func NewMockIdentityProvider() *MockIdentityProvider {
	nacimiento := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}
	return &MockIdentityProvider{ciudadanos: map[string]ciudadanoMock{
		"0926687856": {"2019-05-14", "V4443V4442", domain.IdentityResult{
			Cedula: "0926687856", Nombres: "MARIA JOSE", Apellidos: "PEREZ LOPEZ", Genero: "FEMENINO",
			FechaNacimiento: nacimiento("1985-03-21"), Nacionalidad: "ECUATORIANA", EstadoCivil: "CASADO",
		}},
		"1710034065": {"2018-11-02", "E3331I2222", domain.IdentityResult{
			Cedula: "1710034065", Nombres: "LUIS ALBERTO", Apellidos: "TORRES VERA", Genero: "MASCULINO",
			FechaNacimiento: nacimiento("1958-07-09"), Nacionalidad: "ECUATORIANA", EstadoCivil: "CASADO",
		}},
		"0912345675": {"2020-01-30", "V1111V1111", domain.IdentityResult{
			Cedula: "0912345675", Nombres: "CARLOS ANDRES", Apellidos: "MORA SALAZAR", Genero: "MASCULINO",
			FechaNacimiento: nacimiento("1975-12-01"), Nacionalidad: "ECUATORIANA", EstadoCivil: "SOLTERO",
		}},
		"1304567892": {"2021-08-17", "E2222V3333", domain.IdentityResult{
			Cedula: "1304567892", Nombres: "ANA LUCIA", Apellidos: "GOMEZ RUIZ", Genero: "FEMENINO",
			FechaNacimiento: nacimiento("1990-04-15"), Nacionalidad: "COLOMBIANA", EstadoCivil: "SOLTERO",
		}},
	}}
}

func (m *MockIdentityProvider) Validate(ctx context.Context, cedula string, tipo domain.TipoValidacion, valor string) (*domain.IdentityResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !tipo.EsValido() {
		return nil, domain.ErrInvalidTipoValidacion
	}
	if cedula == CedulaServicioCaido {
		return nil, domain.ErrExternalUnavailable
	}

	c, ok := m.ciudadanos[cedula]
	if !ok {
		return &domain.IdentityResult{Match: false, Cedula: cedula}, nil
	}

	esperado := c.fechaExpedicion
	if tipo == domain.ValidacionCodigoDactilar {
		esperado = c.codigoDactilar
	}
	if valor != esperado {
		// No se revelan datos biográficos si no hay coincidencia (RN-08).
		return &domain.IdentityResult{Match: false, Cedula: cedula}, nil
	}

	res := c.datos
	res.Match = true
	return &res, nil
}
