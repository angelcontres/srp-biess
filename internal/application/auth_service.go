package application

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/security"
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

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "srp_biess_dev_secret_key_change_in_prod"
	}
	return []byte(secret)
}

// GenerarAccessToken crea un Access Token JWT de corta duración (15 minutos) conforme a ADR-006.
func (s *AuthService) GenerarAccessToken(postulanteID, cedula, email string) (string, error) {
	claims := TokenClaims{
		PostulanteID: postulanteID,
		Cedula:       cedula,
		Email:        email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "srp-biess-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(getJWTSecret())
	if err != nil {
		return "", fmt.Errorf("error firmando token: %w", err)
	}

	return tokenStr, nil
}

// GenerarJWT se mantiene por retrocompatibilidad delegando a GenerarAccessToken.
func (s *AuthService) GenerarJWT(postulanteID, cedula, email string) (string, error) {
	return s.GenerarAccessToken(postulanteID, cedula, email)
}

// ValidarAccessToken parsea, verifica la firma HMAC y valida la expiración del Access Token.
func (s *AuthService) ValidarAccessToken(tokenStr string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", t.Header["alg"])
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("claims de token inválidos")
	}

	return claims, nil
}

// GenerarRefreshToken crea un token opaco de alta entropía (32 bytes) para rotación en cookie HttpOnly (ADR-006).
func (s *AuthService) GenerarRefreshToken() (string, error) {
	return security.GenerateOpaqueToken()
}

// HashPassword deriva la contraseña usando Argon2id (ADR-006).
func (s *AuthService) HashPassword(password string) (string, error) {
	return security.HashPassword(password)
}

// VerifyPassword valida la contraseña en texto plano contra el hash Argon2id en tiempo constante.
func (s *AuthService) VerifyPassword(password, hash string) (bool, error) {
	return security.VerifyPassword(password, hash)
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
