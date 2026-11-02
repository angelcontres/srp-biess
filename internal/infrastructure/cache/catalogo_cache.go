package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/redis/go-redis/v9"

	"srp-biess/internal/domain"
)

const (
	DefaultCatalogTTL = 24 * time.Hour
	catalogJitter     = 5 * time.Minute
)

// CachedCatalogoRepository aplica cache-aside sobre el repositorio de catálogos (ADR-005).
// Claves: srp:{env}:cat:{recurso}[:{id}]:v{version}. Si Redis falla, degrada a la BD (RNF-DIS-04).
type CachedCatalogoRepository struct {
	next domain.CatalogoRepository
	rdb  *redis.Client
	env  string
	ttl  time.Duration
}

var _ domain.CatalogoRepository = (*CachedCatalogoRepository)(nil)

func NewCachedCatalogoRepository(next domain.CatalogoRepository, rdb *redis.Client, env string, ttl time.Duration) *CachedCatalogoRepository {
	if ttl <= 0 {
		ttl = DefaultCatalogTTL
	}
	return &CachedCatalogoRepository{next: next, rdb: rdb, env: env, ttl: ttl}
}

func (c *CachedCatalogoRepository) GetPaises(ctx context.Context) ([]domain.Pais, error) {
	return cacheAside(ctx, c, "paises", func() ([]domain.Pais, error) {
		return c.next.GetPaises(ctx)
	})
}

func (c *CachedCatalogoRepository) GetProvincias(ctx context.Context, paisID string) ([]domain.Provincia, error) {
	return cacheAside(ctx, c, "provincias:"+paisID, func() ([]domain.Provincia, error) {
		return c.next.GetProvincias(ctx, paisID)
	})
}

func (c *CachedCatalogoRepository) GetCantones(ctx context.Context, provinciaID string) ([]domain.Canton, error) {
	return cacheAside(ctx, c, "cantones:"+provinciaID, func() ([]domain.Canton, error) {
		return c.next.GetCantones(ctx, provinciaID)
	})
}

// Los cargos no forman parte del alcance de caché de SRP-007.
func (c *CachedCatalogoRepository) GetCargos(ctx context.Context) ([]domain.Cargo, error) {
	return c.next.GetCargos(ctx)
}

func (c *CachedCatalogoRepository) GetCargoByID(ctx context.Context, id string) (*domain.Cargo, error) {
	return c.next.GetCargoByID(ctx, id)
}

// Invalidate incrementa la versión global; las claves anteriores expiran solas por TTL.
func (c *CachedCatalogoRepository) Invalidate(ctx context.Context) error {
	return c.rdb.Incr(ctx, c.versionKey()).Err()
}

func (c *CachedCatalogoRepository) versionKey() string {
	return fmt.Sprintf("srp:%s:cat:version", c.env)
}

func (c *CachedCatalogoRepository) key(ctx context.Context, recurso string) (string, error) {
	v, err := c.rdb.Get(ctx, c.versionKey()).Int64()
	if errors.Is(err, redis.Nil) {
		// Se persiste la versión inicial para que Invalidate (INCR) pase a v2.
		v, err = 1, c.rdb.SetNX(ctx, c.versionKey(), 1, 0).Err()
	}
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("srp:%s:cat:%s:v%d", c.env, recurso, v), nil
}

// ttlConJitter evita la expiración masiva simultánea de claves (SDD 4: 24 h ± 5 min).
func (c *CachedCatalogoRepository) ttlConJitter() time.Duration {
	return c.ttl + time.Duration(rand.Int64N(int64(2*catalogJitter))) - catalogJitter
}

func cacheAside[T any](ctx context.Context, c *CachedCatalogoRepository, recurso string, load func() ([]T, error)) ([]T, error) {
	key, err := c.key(ctx, recurso)
	if err != nil {
		log.Printf("[AVISO] Redis no disponible, consultando BD (%s): %v", recurso, err)
		return load()
	}

	if raw, err := c.rdb.Get(ctx, key).Bytes(); err == nil {
		var items []T
		if err := json.Unmarshal(raw, &items); err == nil {
			return items, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("[AVISO] Redis no disponible, consultando BD (%s): %v", recurso, err)
		return load()
	}

	items, err := load()
	if err != nil {
		return nil, err
	}
	if raw, err := json.Marshal(items); err == nil {
		if err := c.rdb.Set(ctx, key, raw, c.ttlConJitter()).Err(); err != nil {
			log.Printf("[AVISO] No se pudo guardar en caché %s: %v", key, err)
		}
	}
	return items, nil
}
