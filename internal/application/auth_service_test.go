package application_test

import (
	"strings"
	"testing"
	"time"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
)

func TestAuthService_ValidarIdentidad_StrategyPattern(t *testing.T) {
	t.Run("Validación con estrategia por defecto (Cédula Ecuador)", func(t *testing.T) {
		svc := application.NewAuthService(nil, nil)

		// Cédula válida Módulo 10
		if err := svc.ValidarIdentidad("0926687856"); err != nil {
			t.Errorf("esperaba éxito con cédula válida, error: %v", err)
		}

		// Cédula inválida
		if err := svc.ValidarIdentidad("0926687850"); err != domain.ErrInvalidCedula {
			t.Errorf("esperaba ErrInvalidCedula, obtenido: %v", err)
		}
	})

	t.Run("Intercambio dinámico a estrategia Pasaporte", func(t *testing.T) {
		svc := application.NewAuthService(nil, domain.NewCedulaEcuatorianaStrategy())

		// Como tiene estrategia cédula, un pasaporte alfanumérico falla
		if err := svc.ValidarIdentidad("PASAPORTE123"); err == nil {
			t.Errorf("esperaba rechazo de pasaporte con validador de cédula")
		}

		// Cambiamos la estrategia dinámicamente a Pasaporte (Patrón Strategy en acción)
		svc.SetValidator(domain.NewPasaporteStrategy())

		if err := svc.ValidarIdentidad("PASAPORTE123"); err != nil {
			t.Errorf("esperaba éxito al usar estrategia de pasaporte, error: %v", err)
		}
	})
}

func TestAuthService_JWT_And_Tokens(t *testing.T) {
	svc := application.NewAuthService(nil, nil)

	postulanteID := "usr-123456"
	cedula := "0926687856"
	email := "postulante@biess.gob.ec"

	t.Run("Generación y validación exitosa de Access Token (15 min)", func(t *testing.T) {
		tokenStr, err := svc.GenerarAccessToken(postulanteID, cedula, email)
		if err != nil {
			t.Fatalf("error generando access token: %v", err)
		}
		if tokenStr == "" {
			t.Fatalf("token vacío")
		}

		claims, err := svc.ValidarAccessToken(tokenStr)
		if err != nil {
			t.Fatalf("error validando token: %v", err)
		}

		if claims.PostulanteID != postulanteID {
			t.Errorf("postulante_id = %s; want %s", claims.PostulanteID, postulanteID)
		}
		if claims.Cedula != cedula {
			t.Errorf("cedula = %s; want %s", claims.Cedula, cedula)
		}
		if claims.Email != email {
			t.Errorf("email = %s; want %s", claims.Email, email)
		}

		// Verificar que la expiración no supera ~15 minutos
		diff := time.Until(claims.ExpiresAt.Time)
		if diff <= 0 || diff > 16*time.Minute {
			t.Errorf("tiempo de expiración inesperado: %v (debe ser ~15 min)", diff)
		}
	})

	t.Run("Rechazo de token manipulado o corrupto", func(t *testing.T) {
		tokenStr, _ := svc.GenerarAccessToken(postulanteID, cedula, email)
		corruptToken := tokenStr + "tampered"

		_, err := svc.ValidarAccessToken(corruptToken)
		if err == nil {
			t.Errorf("se esperaba error al validar token manipulado")
		}
	})

	t.Run("Generación de Refresh Token opaco", func(t *testing.T) {
		refreshToken, err := svc.GenerarRefreshToken()
		if err != nil {
			t.Fatalf("error generando refresh token: %v", err)
		}
		if len(refreshToken) != 64 {
			t.Errorf("longitud esperada 64 caracteres hex, obtenida: %d", len(refreshToken))
		}
	})

	t.Run("Hash y Verificación con Argon2id", func(t *testing.T) {
		password := "ClaveSegura2026*"
		hash, err := svc.HashPassword(password)
		if err != nil {
			t.Fatalf("error en hash: %v", err)
		}
		if !strings.HasPrefix(hash, "$argon2id$") {
			t.Errorf("el hash no tiene prefijo $argon2id$: %s", hash)
		}

		match, err := svc.VerifyPassword(password, hash)
		if err != nil || !match {
			t.Errorf("se esperaba match con la contraseña correcta")
		}

		wrongMatch, err := svc.VerifyPassword("WrongPassword", hash)
		if err != nil || wrongMatch {
			t.Errorf("no debía coincidir con contraseña equivocada")
		}
	})
}
