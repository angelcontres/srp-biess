package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"srp-biess/internal/domain"
)

// registerFailureScript incrementa el contador y crea el bloqueo en una sola operación atómica (SDD 10.3).
// KEYS[1] = contador, KEYS[2] = bloqueo; ARGV[1] = máximo de intentos, ARGV[2] = TTL en segundos.
// Devuelve el TTL restante del bloqueo en segundos, o 0 si aún no se alcanza el máximo.
// Al crear el bloqueo, ambas claves reciben el mismo TTL y expiran juntas: el contador se reinicia solo.
var registerFailureScript = redis.NewScript(`
local lock_ttl = redis.call('TTL', KEYS[2])
if lock_ttl > 0 then
  return lock_ttl
end
local attempts = redis.call('INCR', KEYS[1])
if attempts == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[2])
end
if attempts >= tonumber(ARGV[1]) then
  redis.call('SET', KEYS[2], attempts, 'EX', ARGV[2])
  redis.call('EXPIRE', KEYS[1], ARGV[2])
  return tonumber(ARGV[2])
end
return 0
`)

// RedisAttemptLimiter implementa domain.AttemptLimiter sobre Redis (SRP-006, ADR-005).
// Claves: srp:{env}:auth:{reg|login}:{attempts|lock}:{cedula} con TTL de 600 s (SDD 4).
type RedisAttemptLimiter struct {
	rdb *redis.Client
	env string
	max int
	ttl time.Duration
}

var _ domain.AttemptLimiter = (*RedisAttemptLimiter)(nil)

func NewRedisAttemptLimiter(rdb *redis.Client, env string) *RedisAttemptLimiter {
	return &RedisAttemptLimiter{rdb: rdb, env: env, max: domain.MaxFailedAttempts, ttl: domain.LockoutDuration}
}

func (l *RedisAttemptLimiter) key(scope domain.AttemptScope, tipo, cedula string) string {
	return fmt.Sprintf("srp:%s:auth:%s:%s:%s", l.env, scope, tipo, cedula)
}

func (l *RedisAttemptLimiter) CheckLock(ctx context.Context, scope domain.AttemptScope, cedula string) (time.Duration, error) {
	ttl, err := l.rdb.TTL(ctx, l.key(scope, "lock", cedula)).Result()
	if err != nil {
		return 0, fmt.Errorf("error consultando bloqueo en Redis: %w", err)
	}
	// TTL negativo: la clave no existe (sin bloqueo).
	if ttl < 0 {
		return 0, nil
	}
	return ttl, nil
}

func (l *RedisAttemptLimiter) RegisterFailure(ctx context.Context, scope domain.AttemptScope, cedula string) (time.Duration, error) {
	keys := []string{l.key(scope, "attempts", cedula), l.key(scope, "lock", cedula)}
	secs, err := registerFailureScript.Run(ctx, l.rdb, keys, l.max, int(l.ttl.Seconds())).Int64()
	if err != nil {
		return 0, fmt.Errorf("error registrando intento fallido en Redis: %w", err)
	}
	return time.Duration(secs) * time.Second, nil
}

func (l *RedisAttemptLimiter) Reset(ctx context.Context, scope domain.AttemptScope, cedula string) error {
	return l.rdb.Del(ctx, l.key(scope, "attempts", cedula), l.key(scope, "lock", cedula)).Err()
}
