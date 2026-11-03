package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrSessionExpired     = errors.New("sesión expirada o token de renovación no válido")
	ErrInvalidCredentials = errors.New("sus credenciales de ingreso son incorrectas")
	ErrAccountBlocked     = errors.New("cédula bloqueada temporalmente por exceso de intentos fallidos")
	ErrAuthProtectionDown = errors.New("protección de autenticación no disponible")
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

// ErrIdentityMismatch es el rechazo genérico de validación de identidad: no revela qué dato falló (RN-08).
var ErrIdentityMismatch = errors.New("no fue posible validar su identidad con los datos ingresados")
