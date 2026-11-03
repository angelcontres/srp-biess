package domain

import (
	"context"
	"errors"
	"time"
)

// AttemptScope identifica el flujo cuyos intentos fallidos se controlan (SDD 4: claves auth:reg y auth:login).
type AttemptScope string

const (
	ScopeRegistro AttemptScope = "reg"
	ScopeLogin    AttemptScope = "login"
)

// Política de bloqueo de EF-11 (RN-06, RN-07): 3 fallos consecutivos bloquean 10 minutos.
const (
	MaxFailedAttempts = 3
	LockoutDuration   = 10 * time.Minute
)

// ErrAttemptLimiterUnavailable indica que no se puede controlar los intentos (Redis caído).
// Se rechaza el intento en lugar de permitir intentos ilimitados (fail-closed, SDD 9.5).
var ErrAttemptLimiterUnavailable = errors.New("no es posible procesar su solicitud en este momento, inténtelo más tarde")

// AttemptLimiter define el puerto del control de intentos fallidos por cédula (SRP-006, SDD 10.3).
type AttemptLimiter interface {
	// CheckLock devuelve el tiempo restante del bloqueo, o 0 si la cédula no está bloqueada.
	CheckLock(ctx context.Context, scope AttemptScope, cedula string) (time.Duration, error)
	// RegisterFailure cuenta un fallo de forma atómica y devuelve el bloqueo restante si se alcanzó el máximo.
	RegisterFailure(ctx context.Context, scope AttemptScope, cedula string) (time.Duration, error)
	// Reset reinicia el contador tras un intento exitoso.
	Reset(ctx context.Context, scope AttemptScope, cedula string) error
}

// LockoutError es el rechazo por bloqueo temporal; la API lo traduce a HTTP 423 con Retry-After.
type LockoutError struct {
	Scope      AttemptScope
	RetryAfter time.Duration
}

// Error devuelve el mensaje exacto del catálogo de errores (SRS 10, códigos AUTH_*_BLOQUEADO).
func (e *LockoutError) Error() string {
	if e.Scope == ScopeLogin {
		return "Cuenta bloqueada temporalmente. Intente nuevamente en 10 minutos."
	}
	return "Registro fallido, inténtelo nuevamente en 10 minutos."
}

func (e *LockoutError) Unwrap() error {
	if e.Scope == ScopeLogin {
		return ErrAccountBlocked
	}
	return ErrCedulaBlocked
}
