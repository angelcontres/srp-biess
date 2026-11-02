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

// fakeCatalogoRepo simula PostgreSQL y cuenta las consultas recibidas.
type fakeCatalogoRepo struct {
	llamadas int
}

func (f *fakeCatalogoRepo) GetPaises(ctx context.Context) ([]domain.Pais, error) {
	f.llamadas++
	return []domain.Pais{{ID: "1", Nombre: "Ecuador", Codigo: "EC"}}, nil
}

func (f *fakeCatalogoRepo) GetProvincias(ctx context.Context, paisID string) ([]domain.Provincia, error) {
	f.llamadas++
	return []domain.Provincia{{ID: "9", PaisID: paisID, Nombre: "Guayas", Codigo: "09"}}, nil
}

func (f *fakeCatalogoRepo) GetCantones(ctx context.Context, provinciaID string) ([]domain.Canton, error) {
	f.llamadas++
	return []domain.Canton{{ID: "1", ProvinciaID: provinciaID, Nombre: "Guayaquil"}}, nil
}

func (f *fakeCatalogoRepo) GetCargos(ctx context.Context) ([]domain.Cargo, error) { return nil, nil }

func (f *fakeCatalogoRepo) GetCargoByID(ctx context.Context, id string) (*domain.Cargo, error) {
	return nil, domain.ErrNotFound
}

func nuevoCache(t *testing.T) (*cache.CachedCatalogoRepository, *fakeCatalogoRepo, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	db := &fakeCatalogoRepo{}
	return cache.NewCachedCatalogoRepository(db, rdb, "dev", cache.DefaultCatalogTTL), db, mr
}

func TestCacheAside_PrimeraConsultaBDLuegoRedis(t *testing.T) {
	repo, db, mr := nuevoCache(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		paises, err := repo.GetPaises(ctx)
		if err != nil || len(paises) != 1 || paises[0].Nombre != "Ecuador" {
			t.Fatalf("GetPaises = %v, %v", paises, err)
		}
	}
	if db.llamadas != 1 {
		t.Errorf("BD consultada %d veces; want 1", db.llamadas)
	}
	if !mr.Exists("srp:dev:cat:paises:v1") {
		t.Errorf("no se guardó la clave srp:dev:cat:paises:v1")
	}
}

func TestCacheAside_TTLConJitter(t *testing.T) {
	repo, _, mr := nuevoCache(t)
	if _, err := repo.GetProvincias(context.Background(), "1"); err != nil {
		t.Fatal(err)
	}
	ttl := mr.TTL("srp:dev:cat:provincias:1:v1")
	if ttl < 24*time.Hour-5*time.Minute || ttl > 24*time.Hour+5*time.Minute {
		t.Errorf("TTL = %v; want 24h ± 5m", ttl)
	}
}

func TestCacheAside_ExpiraYRecarga(t *testing.T) {
	repo, db, mr := nuevoCache(t)
	ctx := context.Background()

	repo.GetCantones(ctx, "9")
	mr.FastForward(25 * time.Hour)
	repo.GetCantones(ctx, "9")

	if db.llamadas != 2 {
		t.Errorf("tras expirar el TTL la BD debe consultarse de nuevo; llamadas = %d", db.llamadas)
	}
}

func TestCacheAside_InvalidarPorVersion(t *testing.T) {
	repo, db, mr := nuevoCache(t)
	ctx := context.Background()

	repo.GetPaises(ctx)
	if err := repo.Invalidate(ctx); err != nil {
		t.Fatal(err)
	}
	repo.GetPaises(ctx)

	if db.llamadas != 2 {
		t.Errorf("tras invalidar la BD debe consultarse de nuevo; llamadas = %d", db.llamadas)
	}
	if !mr.Exists("srp:dev:cat:paises:v2") {
		t.Errorf("no se creó la clave versionada v2")
	}
}

func TestCacheAside_FallbackSiRedisCae(t *testing.T) {
	repo, db, mr := nuevoCache(t)
	mr.Close() // Redis detenido

	paises, err := repo.GetPaises(context.Background())
	if err != nil {
		t.Fatalf("la API no debe fallar si Redis cae: %v", err)
	}
	if len(paises) != 1 || db.llamadas != 1 {
		t.Errorf("debe degradar a la BD; paises=%v llamadas=%d", paises, db.llamadas)
	}
}
