package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/cache"
)

const cedulaPrueba = "0926687856"

func nuevoAttemptStore(t *testing.T) (*cache.RedisAuthAttemptStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	// Sin reintentos para que la prueba de Redis caído no espere el backoff del cliente.
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1})
	t.Cleanup(func() { rdb.Close() })
	return cache.NewRedisAuthAttemptStore(rdb), mr
}

// fallar registra n intentos fallidos y devuelve el resultado del último.
func fallar(t *testing.T, s *cache.RedisAuthAttemptStore, flow domain.AuthFlow, cedula string, n int) (bool, time.Duration) {
	t.Helper()
	var bloqueado bool
	var retry time.Duration
	for i := 0; i < n; i++ {
		var err error
		bloqueado, retry, err = s.RecordFailure(context.Background(), flow, cedula)
		if err != nil {
			t.Fatalf("RecordFailure: %v", err)
		}
	}
	return bloqueado, retry
}

func TestAuthAttempts_TercerFalloBloqueaDiezMinutos(t *testing.T) {
	s, _ := nuevoAttemptStore(t)
	ctx := context.Background()

	if bloqueado, _ := fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 2); bloqueado {
		t.Fatal("dos fallos no deben bloquear")
	}
	if ttl, err := s.CheckBlocked(ctx, domain.AuthFlowLogin, cedulaPrueba); err != nil || ttl != 0 {
		t.Fatalf("antes del tercer fallo: ttl=%v err=%v", ttl, err)
	}

	bloqueado, retry := fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 1)
	if !bloqueado || retry != 10*time.Minute {
		t.Fatalf("tercer fallo: bloqueado=%v retry=%v, se esperaba true y 10m", bloqueado, retry)
	}
	ttl, err := s.CheckBlocked(ctx, domain.AuthFlowLogin, cedulaPrueba)
	if err != nil || ttl <= 0 || ttl > 10*time.Minute {
		t.Fatalf("CheckBlocked tras bloqueo: ttl=%v err=%v", ttl, err)
	}
}

func TestAuthAttempts_FalloDuranteBloqueoNoLoExtiende(t *testing.T) {
	s, mr := nuevoAttemptStore(t)
	fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 3)

	mr.FastForward(4 * time.Minute)
	bloqueado, retry := fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 1)
	if !bloqueado || retry != 6*time.Minute {
		t.Fatalf("fallo durante el bloqueo: bloqueado=%v retry=%v, se esperaba true y 6m", bloqueado, retry)
	}
}

func TestAuthAttempts_BloqueoExpiraALosDiezMinutos(t *testing.T) {
	s, mr := nuevoAttemptStore(t)
	fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 3)

	mr.FastForward(10*time.Minute + time.Second)
	ttl, err := s.CheckBlocked(context.Background(), domain.AuthFlowLogin, cedulaPrueba)
	if err != nil || ttl != 0 {
		t.Fatalf("tras 10 min: ttl=%v err=%v, se esperaba desbloqueado", ttl, err)
	}
	if bloqueado, _ := fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 1); bloqueado {
		t.Fatal("tras expirar, el contador debe empezar de cero")
	}
}

func TestAuthAttempts_VentanaDeIntentosExpira(t *testing.T) {
	s, mr := nuevoAttemptStore(t)
	fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 2)

	mr.FastForward(10*time.Minute + time.Second)
	if bloqueado, _ := fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 1); bloqueado {
		t.Fatal("los fallos de hace más de 10 min no deben sumar")
	}
}

func TestAuthAttempts_ResetLimpiaContadorYBloqueo(t *testing.T) {
	s, _ := nuevoAttemptStore(t)
	ctx := context.Background()

	fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 2)
	if err := s.Reset(ctx, domain.AuthFlowLogin, cedulaPrueba); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if bloqueado, _ := fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 2); bloqueado {
		t.Fatal("tras Reset, dos fallos no deben bloquear")
	}

	fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 1)
	if err := s.Reset(ctx, domain.AuthFlowLogin, cedulaPrueba); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if ttl, _ := s.CheckBlocked(ctx, domain.AuthFlowLogin, cedulaPrueba); ttl != 0 {
		t.Fatalf("Reset debe quitar el bloqueo, ttl=%v", ttl)
	}
}

func TestAuthAttempts_FlujosYCedulasIndependientes(t *testing.T) {
	s, _ := nuevoAttemptStore(t)
	ctx := context.Background()
	fallar(t, s, domain.AuthFlowLogin, cedulaPrueba, 3)

	if ttl, _ := s.CheckBlocked(ctx, domain.AuthFlowRegistration, cedulaPrueba); ttl != 0 {
		t.Fatal("el bloqueo de login no debe afectar al registro")
	}
	if ttl, _ := s.CheckBlocked(ctx, domain.AuthFlowLogin, "1710034065"); ttl != 0 {
		t.Fatal("el bloqueo de una cédula no debe afectar a otra")
	}
}

// Si Redis no responde, el store devuelve error para que el servicio falle cerrado (503).
func TestAuthAttempts_RedisCaidoDevuelveError(t *testing.T) {
	s, mr := nuevoAttemptStore(t)
	ctx := context.Background()
	mr.Close()

	if _, err := s.CheckBlocked(ctx, domain.AuthFlowLogin, cedulaPrueba); err == nil {
		t.Error("CheckBlocked con Redis caído debe devolver error")
	}
	if _, _, err := s.RecordFailure(ctx, domain.AuthFlowLogin, cedulaPrueba); err == nil {
		t.Error("RecordFailure con Redis caído debe devolver error")
	}
	if err := s.Reset(ctx, domain.AuthFlowLogin, cedulaPrueba); err == nil {
		t.Error("Reset con Redis caído debe devolver error")
	}
}

func TestAuthAttempts_SinClienteDevuelveError(t *testing.T) {
	s := cache.NewRedisAuthAttemptStore(nil)
	ctx := context.Background()

	if _, err := s.CheckBlocked(ctx, domain.AuthFlowLogin, cedulaPrueba); err == nil {
		t.Error("CheckBlocked sin cliente debe devolver error")
	}
	if _, _, err := s.RecordFailure(ctx, domain.AuthFlowLogin, cedulaPrueba); err == nil {
		t.Error("RecordFailure sin cliente debe devolver error")
	}
	if err := s.Reset(ctx, domain.AuthFlowLogin, cedulaPrueba); err == nil {
		t.Error("Reset sin cliente debe devolver error")
	}
}
