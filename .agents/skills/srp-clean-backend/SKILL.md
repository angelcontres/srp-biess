---
name: srp-clean-backend
description: >-
  Use this skill whenever implementing, modifying, or refactoring Go Fiber backend code for the SRP-BIESS project. Enforces Clean Architecture (ADR-003), dependency injection, and pgxpool standards.
---

# Clean Backend Implementation Skill (Go Fiber)

Guía operativa para construir componentes backend del SRP-BIESS cumpliendo la rúbrica del examen.

## 1. Estructura de Paquetes
```text
cmd/api/main.go               # Composition Root (Inyección manual)
internal/
├── domain/                  # Entidades y Puertos (Interfaces PURAS)
├── application/             # Casos de uso y lógica de negocio
├── infrastructure/          # Adaptadores (Postgres pgxpool, Redis, SMTP)
└── interfaces/http/         # Handlers Fiber, Rutas y DTOs
```

## 2. Regla Inquebrantable de Dominio
En `internal/domain/`:
* ❌ Prohibido: `import "database/sql"` o `import "github.com/jackc/pgx/v5"`
* ❌ Prohibido: `import "github.com/gofiber/fiber/v2"`
* ✅ Solo tipos primitivos, structs de entidades y contratos de repositorios:
```go
package domain

type CandidatoRepository interface {
    GetByID(ctx context.Context, id string) (*Candidato, error)
    Save(ctx context.Context, c *Candidato) error
}
```

## 3. Verificación de Compilación
Tras cualquier cambio:
```bash
go build ./cmd/api
```
