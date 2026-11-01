package application

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"srp-biess/internal/domain"
)

type AuthService struct {
	repo      domain.PostulanteRepository
	validator domain.DocumentValidationStrategy
}

// NewAuthService construye el servicio inyectando la estrategia de validación (Patrón Strategy + Inyección Manual).
func NewAuthService(repo domain.PostulanteRepository, validator domain.DocumentValidationStrategy) *AuthService {
	if validator == nil {
		validator = domain.NewCedulaEcuatorianaStrategy() // Estrategia por defecto: Módulo 10 ecuatoriano
	}
	return &AuthService{
		repo:      repo,
		validator: validator,
	}
}

type TokenClaims struct {
	PostulanteID string `json:"postulante_id"`
	Cedula       string `json:"cedula"`
	Email        string `json:"email"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerarJWT(postulanteID, cedula, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "srp_biess_dev_secret_key_change_in_prod"
	}

	claims := TokenClaims{
		PostulanteID: postulanteID,
		Cedula:       cedula,
		Email:        email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "srp-biess-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error firmando token: %w", err)
	}

	return tokenStr, nil
}

// ValidarIdentidad delega la validación a la estrategia inyectada (SRP + Strategy Pattern).
func (s *AuthService) ValidarIdentidad(documento string) error {
	valido, err := s.validator.Validate(documento)
	if err != nil || !valido {
		return domain.ErrInvalidCedula
	}
	return nil
}

// SetValidator permite intercambiar la estrategia de validación dinámicamente.
func (s *AuthService) SetValidator(validator domain.DocumentValidationStrategy) {
	if validator != nil {
		s.validator = validator
	}
}
