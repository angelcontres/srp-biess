package application

import (
	"context"
	"log"

	"srp-biess/internal/domain"
)

// WithAttemptLimiter inyecta el control de intentos fallidos en Redis (SRP-006, ADR-003).
func (s *AuthService) WithAttemptLimiter(limiter domain.AttemptLimiter) *AuthService {
	s.limiter = limiter
	return s
}

// verificarBloqueo rechaza el intento si la cédula está bloqueada. Si Redis no responde,
// rechaza con ErrAttemptLimiterUnavailable para no permitir intentos ilimitados (SDD 9.5).
func (s *AuthService) verificarBloqueo(ctx context.Context, scope domain.AttemptScope, cedula string) error {
	if s.limiter == nil {
		return nil
	}
	restante, err := s.limiter.CheckLock(ctx, scope, cedula)
	if err != nil {
		log.Printf("[ALERTA] Control de intentos no disponible (%s): %v", scope, err)
		return domain.ErrAttemptLimiterUnavailable
	}
	if restante > 0 {
		return &domain.LockoutError{Scope: scope, RetryAfter: restante}
	}
	return nil
}

// registrarFallo cuenta el fallo y devuelve el bloqueo si fue el último intento permitido;
// en otro caso devuelve el error original del intento.
func (s *AuthService) registrarFallo(ctx context.Context, scope domain.AttemptScope, cedula string, fallo error) error {
	if s.limiter == nil {
		return fallo
	}
	restante, err := s.limiter.RegisterFailure(ctx, scope, cedula)
	if err != nil {
		log.Printf("[ALERTA] Control de intentos no disponible (%s): %v", scope, err)
		return domain.ErrAttemptLimiterUnavailable
	}
	if restante > 0 {
		return &domain.LockoutError{Scope: scope, RetryAfter: restante}
	}
	return fallo
}

// reiniciarIntentos limpia el contador tras un intento exitoso (SDD 9.5: los fallos deben ser consecutivos).
func (s *AuthService) reiniciarIntentos(ctx context.Context, scope domain.AttemptScope, cedula string) {
	if s.limiter == nil {
		return
	}
	if err := s.limiter.Reset(ctx, scope, cedula); err != nil {
		log.Printf("[AVISO] No se pudo reiniciar el contador de intentos (%s): %v", scope, err)
	}
}
