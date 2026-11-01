package application

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"srp-biess/internal/domain"
)

type AuthService struct {
	repo domain.PostulanteRepository
}

func NewAuthService(repo domain.PostulanteRepository) *AuthService {
	return &AuthService{repo: repo}
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

func (s *AuthService) ValidarIdentidad(cedula string) error {
	if !domain.ValidarCedulaEcuador(cedula) {
		return domain.ErrInvalidCedula
	}
	return nil
}
