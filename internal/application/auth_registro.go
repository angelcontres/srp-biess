package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"srp-biess/internal/domain"
)

// TokenValidacionTTL es la vigencia del token que habilita el paso 2 del registro (SDD 7.4.2).
const TokenValidacionTTL = 10 * time.Minute

const propositoRegistro = "registro"

// RegistroClaims son los claims del token de validación de identidad.
type RegistroClaims struct {
	Cedula    string `json:"cedula"`
	Proposito string `json:"proposito"`
	jwt.RegisteredClaims
}

// IdentidadValidada es el resultado exitoso de POST /auth/registro/validar-identidad.
type IdentidadValidada struct {
	TokenValidacion string
	Datos           *domain.IdentityResult
}

// WithIdentityProvider inyecta el adaptador del Registro Civil (ADR-003, ADR-010).
func (s *AuthService) WithIdentityProvider(identity domain.IdentityProvider) *AuthService {
	s.identity = identity
	return s
}

// ValidarIdentidadRegistro contrasta cédula + dato con el Registro Civil (SDD 7.4.1).
// Aplica primero el Módulo 10 (SRP-008): si falla, no se llama al proveedor externo.
func (s *AuthService) ValidarIdentidadRegistro(ctx context.Context, cedula string, tipo domain.TipoValidacion, valor string) (*IdentidadValidada, error) {
	if !tipo.EsValido() {
		return nil, domain.ErrInvalidTipoValidacion
	}
	if err := s.ValidarIdentidad(cedula); err != nil {
		return nil, domain.ErrInvalidCedula
	}
	if err := s.checkAuthBlocked(ctx, domain.AuthFlowRegistration, cedula); err != nil {
		return nil, err
	}
	if s.identity == nil {
		return nil, domain.ErrExternalUnavailable
	}

	res, err := s.identity.Validate(ctx, cedula, tipo, valor)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTipoValidacion) {
			return nil, err
		}
		return nil, domain.ErrExternalUnavailable
	}
	if res == nil || !res.Match {
		if recordErr := s.recordAuthFailure(ctx, domain.AuthFlowRegistration, cedula); recordErr != nil {
			return nil, recordErr
		}
		return nil, domain.ErrIdentityMismatch
	}
	if err := s.resetAuthFailures(ctx, domain.AuthFlowRegistration, cedula); err != nil {
		return nil, err
	}

	token, err := s.generarTokenValidacion(cedula)
	if err != nil {
		return nil, err
	}
	return &IdentidadValidada{TokenValidacion: token, Datos: res}, nil
}

func (s *AuthService) generarTokenValidacion(cedula string) (string, error) {
	now := time.Now()
	claims := RegistroClaims{
		Cedula:    cedula,
		Proposito: propositoRegistro,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   cedula,
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenValidacionTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "srp-biess-api",
		},
	}
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(getJWTSecret())
	if err != nil {
		return "", fmt.Errorf("error firmando token de validación: %w", err)
	}
	return tokenStr, nil
}

// ValidarTokenValidacion verifica firma, vigencia y propósito del token (lo usará POST /auth/registro, SDD 7.4.2).
func (s *AuthService) ValidarTokenValidacion(tokenStr string) (*RegistroClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RegistroClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", t.Header["alg"])
		}
		return getJWTSecret(), nil
	})
	if err != nil {
		return nil, fmt.Errorf("token de validación inválido: %w", err)
	}
	claims, ok := token.Claims.(*RegistroClaims)
	if !ok || !token.Valid || claims.Proposito != propositoRegistro {
		return nil, errors.New("token de validación inválido")
	}
	return claims, nil
}
