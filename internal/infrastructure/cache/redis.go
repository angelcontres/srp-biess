package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient inicializa el cliente de Redis con fallback tolerante a fallos.
func NewRedisClient(ctx context.Context) *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	password := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Password:     password,
		DB:           0,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Printf("[AVISO] No se pudo conectar a Redis: %v (degradando a modo sin caché)", err)
	} else {
		log.Println("[INFO] Conexión exitosa a Redis 7")
	}

	return rdb
}

// KeyEnv devuelve el segmento {env} de las claves srp:{env}:... (SDD 4) a partir de APP_ENV.
func KeyEnv() string {
	switch env := os.Getenv("APP_ENV"); env {
	case "", "development":
		return "dev"
	case "production":
		return "prod"
	default:
		return env
	}
}
