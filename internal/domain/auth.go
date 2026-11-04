package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
)

var (
	ErrSessionExpired     = errors.New("sesión expirada o token de renovación no válido")
	ErrInvalidCredentials = errors.New("sus credenciales de ingreso son incorrectas")
	ErrAccountBlocked     = errors.New("cédula bloqueada temporalmente por exceso de intentos fallidos")
	ErrAuthProtectionDown = errors.New("protección de autenticación no disponible")
	ErrCedulaDuplicada    = errors.New("AUTH_CEDULA_DUPLICADA: ya existe una cuenta registrada con esta cédula")
	ErrCorreoDuplicado    = errors.New("AUTH_CORREO_DUPLICADO: ya existe una cuenta registrada con este correo electrónico")
	ErrPasswordInvalida   = errors.New("la contraseña debe tener entre 10 y 128 caracteres e incluir al menos 3 de 4 tipos: mayúsculas, minúsculas, números y caracteres especiales")
	ErrCorreoInvalido     = errors.New("el formato del correo electrónico es inválido")
	ErrTokenInvalido      = errors.New("el token de validación es inválido o ha expirado")
)

type AuthFlow string

const (
	AuthFlowRegistration AuthFlow = "reg"
	AuthFlowLogin        AuthFlow = "login"
)

type AuthBlockedError struct {
	Flow       AuthFlow
	RetryAfter time.Duration
}

func (e *AuthBlockedError) Error() string {
	return fmt.Sprintf("%s: %s", ErrAccountBlocked, e.Flow)
}

func (e *AuthBlockedError) Unwrap() error {
	return ErrAccountBlocked
}

type AuthAttemptStore interface {
	CheckBlocked(ctx context.Context, flow AuthFlow, cedula string) (time.Duration, error)
	RecordFailure(ctx context.Context, flow AuthFlow, cedula string) (bool, time.Duration, error)
	Reset(ctx context.Context, flow AuthFlow, cedula string) error
}

// SessionStore define el puerto abstracto de persistencia para refresh tokens y sesiones (ADR-006).
type SessionStore interface {
	SaveRefreshToken(ctx context.Context, token string, postulanteID string, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, token string) (string, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

// PasswordResetStore define el puerto para tokens de recuperación de contraseña (SDD 10.5).
type PasswordResetStore interface {
	SaveResetToken(ctx context.Context, token string, postulanteID string, ttl time.Duration) error
	GetResetToken(ctx context.Context, token string) (string, error)
	DeleteResetToken(ctx context.Context, token string) error
}

// ErrIdentityMismatch es el rechazo genérico de validación de identidad: no revela qué dato falló (RN-08).
var ErrIdentityMismatch = errors.New("no fue posible validar su identidad con los datos ingresados")

// ValidarPoliticaPassword aplica la política de seguridad de SDD 9.2.2:
// - Longitud mínima 10, máxima 128 caracteres.
// - Al menos 3 de 4 clases: mayúsculas, minúsculas, números y símbolos.
// - No puede ser idéntica a la cédula ni al correo.
func ValidarPoliticaPassword(password, cedula, correo string) error {
	trimmed := strings.TrimSpace(password)
	if len(trimmed) < 10 || len(trimmed) > 128 {
		return ErrPasswordInvalida
	}

	if cedula != "" && strings.EqualFold(trimmed, strings.TrimSpace(cedula)) {
		return errors.New("la contraseña no puede ser igual al número de cédula")
	}
	if correo != "" && strings.EqualFold(trimmed, strings.TrimSpace(correo)) {
		return errors.New("la contraseña no puede ser igual al correo electrónico")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range trimmed {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r) || strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?`~", r):
			hasSpecial = true
		}
	}

	clases := 0
	if hasUpper {
		clases++
	}
	if hasLower {
		clases++
	}
	if hasDigit {
		clases++
	}
	if hasSpecial {
		clases++
	}

	if clases < 3 {
		return ErrPasswordInvalida
	}
	return nil
}
