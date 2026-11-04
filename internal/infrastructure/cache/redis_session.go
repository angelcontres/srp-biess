package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"srp-biess/internal/domain"
)

type memoryTokenItem struct {
	postulanteID string
	expiresAt    time.Time
}

// RedisSessionStore implementa domain.SessionStore y domain.PasswordResetStore sobre Redis con fallback seguro en memoria (ADR-005, ADR-006).
type RedisSessionStore struct {
	rdb         *redis.Client
	mu          sync.RWMutex
	memoryStore map[string]memoryTokenItem
}

// NewRedisSessionStore construye el almacén de sesiones y tokens.
func NewRedisSessionStore(rdb *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{
		rdb:         rdb,
		memoryStore: make(map[string]memoryTokenItem),
	}
}

func (s *RedisSessionStore) tokenKey(token string) string {
	return fmt.Sprintf("refresh_token:%s", token)
}

func (s *RedisSessionStore) resetKey(token string) string {
	return fmt.Sprintf("pwd_reset:%s", token)
}

// SaveRefreshToken persiste el token de actualización con TTL de expiración (8 horas).
func (s *RedisSessionStore) SaveRefreshToken(ctx context.Context, token string, postulanteID string, ttl time.Duration) error {
	if s.rdb != nil {
		err := s.rdb.Set(ctx, s.tokenKey(token), postulanteID, ttl).Err()
		if err == nil {
			return nil
		}
	}

	// Fallback concurrente en memoria
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memoryStore[token] = memoryTokenItem{
		postulanteID: postulanteID,
		expiresAt:    time.Now().Add(ttl),
	}
	return nil
}

// GetRefreshToken obtiene el postulante asociado al token activo.
func (s *RedisSessionStore) GetRefreshToken(ctx context.Context, token string) (string, error) {
	if s.rdb != nil {
		val, err := s.rdb.Get(ctx, s.tokenKey(token)).Result()
		if err == nil {
			return val, nil
		}
		if err == redis.Nil {
			return "", domain.ErrSessionExpired
		}
	}

	// Fallback concurrente en memoria
	s.mu.RLock()
	item, ok := s.memoryStore[token]
	s.mu.RUnlock()

	if !ok || time.Now().After(item.expiresAt) {
		if ok {
			s.mu.Lock()
			delete(s.memoryStore, token)
			s.mu.Unlock()
		}
		return "", domain.ErrSessionExpired
	}

	return item.postulanteID, nil
}

// DeleteRefreshToken revoca el token inmediatamente (ADR-006 / Logout / Rotación).
func (s *RedisSessionStore) DeleteRefreshToken(ctx context.Context, token string) error {
	if s.rdb != nil {
		_ = s.rdb.Del(ctx, s.tokenKey(token)).Err()
	}

	s.mu.Lock()
	delete(s.memoryStore, token)
	s.mu.Unlock()
	return nil
}

// SaveResetToken persiste el token de recuperación de contraseña con TTL (30 min por defecto, SDD 10.5).
func (s *RedisSessionStore) SaveResetToken(ctx context.Context, token string, postulanteID string, ttl time.Duration) error {
	if s.rdb != nil {
		err := s.rdb.Set(ctx, s.resetKey(token), postulanteID, ttl).Err()
		if err == nil {
			return nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.memoryStore["reset:"+token] = memoryTokenItem{
		postulanteID: postulanteID,
		expiresAt:    time.Now().Add(ttl),
	}
	return nil
}

// GetResetToken recupera el postulanteID si el token es válido y no ha expirado.
func (s *RedisSessionStore) GetResetToken(ctx context.Context, token string) (string, error) {
	if s.rdb != nil {
		val, err := s.rdb.Get(ctx, s.resetKey(token)).Result()
		if err == nil {
			return val, nil
		}
		if err == redis.Nil {
			return "", domain.ErrTokenInvalido
		}
	}

	s.mu.RLock()
	item, ok := s.memoryStore["reset:"+token]
	s.mu.RUnlock()

	if !ok || time.Now().After(item.expiresAt) {
		if ok {
			s.mu.Lock()
			delete(s.memoryStore, "reset:"+token)
			s.mu.Unlock()
		}
		return "", domain.ErrTokenInvalido
	}

	return item.postulanteID, nil
}

// DeleteResetToken invalida el token de recuperación tras haber sido utilizado.
func (s *RedisSessionStore) DeleteResetToken(ctx context.Context, token string) error {
	if s.rdb != nil {
		_ = s.rdb.Del(ctx, s.resetKey(token)).Err()
	}

	s.mu.Lock()
	delete(s.memoryStore, "reset:"+token)
	s.mu.Unlock()
	return nil
}
