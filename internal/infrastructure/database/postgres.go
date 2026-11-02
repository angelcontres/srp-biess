package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool inicializa el pool de conexiones de PostgreSQL con pgxpool.
// El tamaño del pool y los tiempos de vida se configuran mediante variables de entorno (SRP-004, ADR-004).
func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "srp_user")
	password := getEnvOrDefault("DB_PASSWORD", "srp_secret_password")
	dbname := getEnvOrDefault("DB_NAME", "srp_biess")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parseando configuración pgxpool: %w", err)
	}

	// Criterio de Aceptación 1: Conexiones concurrentes controladas mediante parámetros de entorno
	maxConns := getEnvIntOrDefault("DB_MAX_CONNS", 25)
	minConns := getEnvIntOrDefault("DB_MIN_CONNS", 5)
	maxConnLifetime := getEnvDurationOrDefault("DB_MAX_CONN_LIFETIME", 30*time.Minute)
	maxConnIdleTime := getEnvDurationOrDefault("DB_MAX_CONN_IDLE_TIME", 5*time.Minute)

	config.MaxConns = int32(maxConns)
	config.MinConns = int32(minConns)
	config.MaxConnLifetime = maxConnLifetime
	config.MaxConnIdleTime = maxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error conectando a postgres: %w", err)
	}

	// Ping con timeout para verificar salud inicial
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		log.Printf("[AVISO] No se pudo hacer ping a PostgreSQL: %v (verifique que Docker esté levantado)", err)
	} else {
		log.Printf("[INFO] Conexión exitosa a PostgreSQL 16 (pgxpool: min=%d, max=%d)", minConns, maxConns)
	}

	return pool, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvIntOrDefault(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			return n
		}
	}
	return defaultVal
}

func getEnvDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil && d > 0 {
			return d
		}
	}
	return defaultVal
}
