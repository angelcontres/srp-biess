package cache_test

import (
	"context"
	"testing"
	"time"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/cache"
)

func TestRedisSessionStore_MemoryFallback(t *testing.T) {
	ctx := context.Background()
	store := cache.NewRedisSessionStore(nil) // nil activa el fallback seguro en memoria

	token := "token-aleatorio-hex-12345"
	postulanteID := "postulante-uuid-999"

	// 1. Guardar y recuperar
	err := store.SaveRefreshToken(ctx, token, postulanteID, 1*time.Hour)
	if err != nil {
		t.Fatalf("error guardando token: %v", err)
	}

	gotID, err := store.GetRefreshToken(ctx, token)
	if err != nil {
		t.Fatalf("error obteniendo token: %v", err)
	}
	if gotID != postulanteID {
		t.Errorf("got %s; want %s", gotID, postulanteID)
	}

	// 2. Revocación / Eliminación
	err = store.DeleteRefreshToken(ctx, token)
	if err != nil {
		t.Fatalf("error eliminando token: %v", err)
	}

	_, err = store.GetRefreshToken(ctx, token)
	if err != domain.ErrSessionExpired {
		t.Errorf("se esperaba ErrSessionExpired tras eliminar token, obtenido: %v", err)
	}

	// 3. Expiración de TTL
	shortToken := "token-corto-expirado"
	_ = store.SaveRefreshToken(ctx, shortToken, postulanteID, 10*time.Millisecond)
	time.Sleep(15 * time.Millisecond)

	_, err = store.GetRefreshToken(ctx, shortToken)
	if err != domain.ErrSessionExpired {
		t.Errorf("se esperaba ErrSessionExpired por tiempo, obtenido: %v", err)
	}
}
