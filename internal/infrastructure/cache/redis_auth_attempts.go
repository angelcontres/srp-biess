package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"srp-biess/internal/domain"
)

const authAttemptTTL = 10 * time.Minute

var recordAuthFailureScript = redis.NewScript(`
local lock_ttl = redis.call("TTL", KEYS[2])
if lock_ttl > 0 then
  return {1, lock_ttl}
end

local attempts = redis.call("INCR", KEYS[1])
if attempts == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[1])
end
if attempts >= tonumber(ARGV[2]) then
  redis.call("SET", KEYS[2], "1", "EX", ARGV[1])
  redis.call("DEL", KEYS[1])
  return {1, ARGV[1]}
end
return {0, 0}
`)

type RedisAuthAttemptStore struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisAuthAttemptStore(rdb *redis.Client) *RedisAuthAttemptStore {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	return &RedisAuthAttemptStore{
		rdb:    rdb,
		prefix: fmt.Sprintf("srp:%s:auth", env),
	}
}

func (s *RedisAuthAttemptStore) keys(flow domain.AuthFlow, cedula string) (string, string) {
	base := fmt.Sprintf("%s:%s", s.prefix, flow)
	return fmt.Sprintf("%s:attempts:%s", base, cedula), fmt.Sprintf("%s:lock:%s", base, cedula)
}

func (s *RedisAuthAttemptStore) CheckBlocked(ctx context.Context, flow domain.AuthFlow, cedula string) (time.Duration, error) {
	if s.rdb == nil {
		return 0, fmt.Errorf("cliente Redis no configurado")
	}
	_, lockKey := s.keys(flow, cedula)
	ttl, err := s.rdb.TTL(ctx, lockKey).Result()
	if err != nil {
		return 0, err
	}
	if ttl <= 0 {
		return 0, nil
	}
	return ttl, nil
}

func (s *RedisAuthAttemptStore) RecordFailure(ctx context.Context, flow domain.AuthFlow, cedula string) (bool, time.Duration, error) {
	if s.rdb == nil {
		return false, 0, fmt.Errorf("cliente Redis no configurado")
	}
	attemptsKey, lockKey := s.keys(flow, cedula)
	values, err := recordAuthFailureScript.Run(ctx, s.rdb, []string{attemptsKey, lockKey}, int(authAttemptTTL/time.Second), 3).Result()
	if err != nil {
		return false, 0, err
	}
	result, ok := values.([]interface{})
	if !ok || len(result) != 2 {
		return false, 0, fmt.Errorf("respuesta inválida del script de intentos")
	}
	blocked, err := redisInt(result[0])
	if err != nil {
		return false, 0, err
	}
	retrySeconds, err := redisInt(result[1])
	if err != nil {
		return false, 0, err
	}
	return blocked == 1, time.Duration(retrySeconds) * time.Second, nil
}

func (s *RedisAuthAttemptStore) Reset(ctx context.Context, flow domain.AuthFlow, cedula string) error {
	if s.rdb == nil {
		return fmt.Errorf("cliente Redis no configurado")
	}
	attemptsKey, lockKey := s.keys(flow, cedula)
	return s.rdb.Del(ctx, attemptsKey, lockKey).Err()
}

func redisInt(value interface{}) (int, error) {
	switch typed := value.(type) {
	case int64:
		return int(typed), nil
	case string:
		return strconv.Atoi(typed)
	case []byte:
		return strconv.Atoi(string(typed))
	default:
		return 0, fmt.Errorf("valor entero Redis inválido: %T", value)
	}
}
