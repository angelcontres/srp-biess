package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool inicializa el pool de conexiones de PostgreSQL con pgxpool.
func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "srp_user"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "srp_secret_password"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "srp_biess"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parseando configuración pgxpool: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error conectando a postgres: %w", err)
	}

	// Ping con timeout
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		log.Printf("[AVISO] No se pudo hacer ping a PostgreSQL: %v (verifique que Docker esté levantado)", err)
	} else {
		log.Println("[INFO] Conexión exitosa a PostgreSQL 16 (pgxpool)")
	}

	return pool, nil
}
