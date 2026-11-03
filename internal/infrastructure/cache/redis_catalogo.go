package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"srp-biess/internal/domain"
)

const catalogoTTL = 24 * time.Hour

type RedisCatalogoCache struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisCatalogoCache(rdb *redis.Client) *RedisCatalogoCache {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	return &RedisCatalogoCache{
		rdb:    rdb,
		prefix: fmt.Sprintf("srp:%s:cat", env),
	}
}

func (c *RedisCatalogoCache) versionKey() string {
	return c.prefix + ":version"
}

func (c *RedisCatalogoCache) version(ctx context.Context) (string, error) {
	value, err := c.rdb.Get(ctx, c.versionKey()).Result()
	if err == redis.Nil {
		if err := c.rdb.Set(ctx, c.versionKey(), "1", 0).Err(); err != nil {
			return "", err
		}
		return "1", nil
	}
	return value, err
}

func (c *RedisCatalogoCache) ttl() time.Duration {
	return catalogoTTL + time.Duration(rand.Int63n(int64(10*time.Minute))) - 5*time.Minute
}

func (c *RedisCatalogoCache) get(ctx context.Context, key string, target interface{}) (bool, error) {
	if c.rdb == nil {
		return false, fmt.Errorf("cliente Redis no configurado")
	}
	value, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal([]byte(value), target)
}

func (c *RedisCatalogoCache) set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, c.ttl()).Err()
}

func (c *RedisCatalogoCache) key(ctx context.Context, resource, id string) (string, error) {
	version, err := c.version(ctx)
	if err != nil {
		return "", err
	}
	if id == "" {
		return fmt.Sprintf("%s:%s:%s", c.prefix, resource, version), nil
	}
	return fmt.Sprintf("%s:%s:%s:%s", c.prefix, resource, id, version), nil
}

func (c *RedisCatalogoCache) GetPaises(ctx context.Context) ([]domain.Pais, bool, error) {
	key, err := c.key(ctx, "paises", "")
	if err != nil {
		return nil, false, err
	}
	var result []domain.Pais
	ok, err := c.get(ctx, key, &result)
	return result, ok, err
}

func (c *RedisCatalogoCache) SetPaises(ctx context.Context, paises []domain.Pais) error {
	key, err := c.key(ctx, "paises", "")
	if err != nil {
		return err
	}
	return c.set(ctx, key, paises)
}

func (c *RedisCatalogoCache) GetProvincias(ctx context.Context, paisID string) ([]domain.Provincia, bool, error) {
	key, err := c.key(ctx, "provincias", paisID)
	if err != nil {
		return nil, false, err
	}
	var result []domain.Provincia
	ok, err := c.get(ctx, key, &result)
	return result, ok, err
}

func (c *RedisCatalogoCache) SetProvincias(ctx context.Context, paisID string, provincias []domain.Provincia) error {
	key, err := c.key(ctx, "provincias", paisID)
	if err != nil {
		return err
	}
	return c.set(ctx, key, provincias)
}

func (c *RedisCatalogoCache) GetCantones(ctx context.Context, provinciaID string) ([]domain.Canton, bool, error) {
	key, err := c.key(ctx, "cantones", provinciaID)
	if err != nil {
		return nil, false, err
	}
	var result []domain.Canton
	ok, err := c.get(ctx, key, &result)
	return result, ok, err
}

func (c *RedisCatalogoCache) SetCantones(ctx context.Context, provinciaID string, cantones []domain.Canton) error {
	key, err := c.key(ctx, "cantones", provinciaID)
	if err != nil {
		return err
	}
	return c.set(ctx, key, cantones)
}

func (c *RedisCatalogoCache) Invalidate(ctx context.Context) error {
	_, err := c.rdb.Incr(ctx, c.versionKey()).Result()
	return err
}
