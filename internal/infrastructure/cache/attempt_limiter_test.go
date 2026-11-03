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

func setupLimiter(t *testing.T) (*miniredis.Miniredis, *cache.RedisAttemptLimiter) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return mr, cache.NewRedisAttemptLimiter(rdb, "dev")
}

func fallar(t *testing.T, l *cache.RedisAttemptLimiter, scope domain.AttemptScope, veces int) time.Duration {
	t.Helper()
	var restante time.Duration
	for i := 0; i < veces; i++ {
		var err error
		if restante, err = l.RegisterFailure(context.Background(), scope, cedulaPrueba); err != nil {
			t.Fatalf("RegisterFailure: %v", err)
		}
	}
	return restante
}

func TestRedisAttemptLimiter(t *testing.T) {
	ctx := context.Background()

	t.Run("el 3er fallo bloquea 600 s con las claves del SDD", func(t *testing.T) {
		mr, l := setupLimiter(t)
		if restante := fallar(t, l, domain.ScopeRegistro, 2); restante != 0 {
			t.Fatalf("bloqueo tras 2 fallos = %v; want 0", restante)
		}
		if restante := fallar(t, l, domain.ScopeRegistro, 1); restante != 600*time.Second {
			t.Fatalf("bloqueo tras 3 fallos = %v; want 600s", restante)
		}
		lockKey := "srp:dev:auth:reg:lock:" + cedulaPrueba
		if ttl := mr.TTL(lockKey); ttl != 600*time.Second {
			t.Errorf("TTL(%s) = %v; want 600s", lockKey, ttl)
		}
		if restante, err := l.CheckLock(ctx, domain.ScopeRegistro, cedulaPrueba); err != nil || restante != 600*time.Second {
			t.Errorf("CheckLock = %v, %v; want 600s", restante, err)
		}
	})

	t.Run("al expirar el TTL el bloqueo y el contador se reinician solos", func(t *testing.T) {
		mr, l := setupLimiter(t)
		fallar(t, l, domain.ScopeRegistro, 3)
		mr.FastForward(600 * time.Second)

		if restante, _ := l.CheckLock(ctx, domain.ScopeRegistro, cedulaPrueba); restante != 0 {
			t.Fatalf("bloqueo tras expirar = %v; want 0", restante)
		}
		if mr.Exists("srp:dev:auth:reg:attempts:" + cedulaPrueba) {
			t.Error("el contador debe expirar junto con el bloqueo")
		}
		if restante := fallar(t, l, domain.ScopeRegistro, 2); restante != 0 {
			t.Errorf("tras expirar, 2 fallos no deben bloquear: %v", restante)
		}
	})

	t.Run("un fallo durante el bloqueo no lo extiende", func(t *testing.T) {
		mr, l := setupLimiter(t)
		fallar(t, l, domain.ScopeRegistro, 3)
		mr.FastForward(200 * time.Second)
		if restante := fallar(t, l, domain.ScopeRegistro, 1); restante != 400*time.Second {
			t.Errorf("bloqueo restante = %v; want 400s", restante)
		}
	})

	t.Run("Reset reinicia el contador: los fallos deben ser consecutivos", func(t *testing.T) {
		_, l := setupLimiter(t)
		fallar(t, l, domain.ScopeLogin, 2)
		if err := l.Reset(ctx, domain.ScopeLogin, cedulaPrueba); err != nil {
			t.Fatalf("Reset: %v", err)
		}
		if restante := fallar(t, l, domain.ScopeLogin, 2); restante != 0 {
			t.Errorf("tras Reset, 2 fallos no deben bloquear: %v", restante)
		}
	})

	t.Run("registro y login se cuentan por separado", func(t *testing.T) {
		_, l := setupLimiter(t)
		fallar(t, l, domain.ScopeLogin, 3)
		if restante, _ := l.CheckLock(ctx, domain.ScopeRegistro, cedulaPrueba); restante != 0 {
			t.Errorf("el bloqueo de login no debe bloquear el registro: %v", restante)
		}
	})

	t.Run("con Redis caído devuelve error (fail-closed)", func(t *testing.T) {
		mr, l := setupLimiter(t)
		mr.Close()
		if _, err := l.CheckLock(ctx, domain.ScopeRegistro, cedulaPrueba); err == nil {
			t.Error("CheckLock debe fallar sin Redis")
		}
		if _, err := l.RegisterFailure(ctx, domain.ScopeRegistro, cedulaPrueba); err == nil {
			t.Error("RegisterFailure debe fallar sin Redis")
		}
	})
}
